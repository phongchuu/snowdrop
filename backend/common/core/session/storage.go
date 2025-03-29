package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"internal.snowdrop/common/core/database"
	"internal.snowdrop/common/store"
)

// PostgresRepository implements Repository for PostgreSQL.
type PostgresRepository struct {
	txMgr database.TransactionManager
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(transactionManager database.TransactionManager) *PostgresRepository {
	return &PostgresRepository{
		txMgr: transactionManager,
	}
}

func (r *PostgresRepository) GetSession(
	ctx context.Context,
	sessionID string,
) (*store.SessionModel, error) {
	return database.RunTxWithData(
		ctx,
		r.txMgr,
		func(ctx context.Context, tx *sql.Tx) (*store.SessionModel, error) {
			var session store.SessionModel

			err := sqlscan.Get(
				ctx,
				tx,
				&session,
				`SELECT id, user_id, created_at, last_accessed_at, expires_at, data
                FROM sessions
                WHERE id = $1 AND expires_at > NOW()
                FOR UPDATE`,
				sessionID,
			)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, nil
				}

				return nil, err
			}

			return &session, nil
		},
	)
}

func (r *PostgresRepository) UpdateSession(
	ctx context.Context,
	sessionID string,
	lastAccessedAt time.Time,
	expiresAt time.Time,
) error {
	return database.RunTx(ctx, r.txMgr, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`UPDATE sessions SET last_accessed_at = $1, expires_at = $2 WHERE id = $3`,
			lastAccessedAt,
			expiresAt,
			sessionID,
		)

		return err
	})
}

func (r *PostgresRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	expiresAt time.Time,
) (*store.SessionModel, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return database.RunTxWithData(
		ctx,
		r.txMgr,
		func(ctx context.Context, tx *sql.Tx) (*store.SessionModel, error) {
			model := store.SessionModel{
				ID: sessionID,
				UserID: uuid.NullUUID{
					UUID:  lo.Ternary(userID == uuid.Nil, uuid.Nil, userID),
					Valid: userID != uuid.Nil,
				},
				CreatedAt:      now,
				LastAccessedAt: now,
				ExpiresAt:      expiresAt,
				Data:           json.RawMessage{},
			}

			if err := sqlscan.Get(
				ctx,
				tx,
				&model,
				`INSERT INTO sessions (id, user_id, created_at, last_accessed_at, expires_at, data)
		    VALUES ($1, $2, $3, $4, $5, '{}'::jsonb)
            RETURNING *`,
				model.ID,
				model.UserID,
				model.CreatedAt,
				model.LastAccessedAt,
				model.ExpiresAt,
			); err != nil {
				return nil, err
			}

			return &model, nil
		},
	)
}

func (r *PostgresRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return database.RunTx(ctx, r.txMgr, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE id = $1", sessionID)

		return err
	})
}

func (r *PostgresRepository) DeleteExpiredSessions(ctx context.Context) error {
	return database.RunTx(ctx, r.txMgr, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")

		return err
	})
}
