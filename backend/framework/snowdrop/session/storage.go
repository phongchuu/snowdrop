package session

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/opentelemetry"
)

// PostgresRepository implements Repository for PostgreSQL.
type PostgresRepository struct {
	structName string
	txMgr      snowdrop.TransactionManager
	tracer     trace.Tracer
}

var _ snowdrop.SessionRepository = (*PostgresRepository)(nil)

func NewPostgresRepository(transactionManager snowdrop.TransactionManager, tracer trace.Tracer) *PostgresRepository {
	repository := PostgresRepository{}
	repository.structName = reflect.TypeOf(repository).Name()
	repository.txMgr = transactionManager
	repository.tracer = tracer

	return &repository
}

func (r *PostgresRepository) GetSession(
	ctx context.Context,
	sessionID string,
) (*snowdrop.SessionModel, error) {
	spanCtx, span := r.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(r.structName, "GetSession"),
		trace.WithAttributes(attribute.String("args[1].sessionID", sessionID)),
	)
	defer span.End()

	return database.RunTxWithData(
		spanCtx,
		r.txMgr,
		func(ctx context.Context, tx *gorm.DB) (*snowdrop.SessionModel, error) {
			var session snowdrop.SessionModel
			result := tx.First(&session, "id = ? AND expires_at > NOW()", sessionID)
			if result.Error != nil {
				return nil, result.Error
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
	spanCtx, span := r.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(r.structName, "UpdateSession"),
		trace.WithAttributes(
			attribute.String("args[1].sessionID", sessionID),
			attribute.Stringer("args[2].lastAccessedAt", lastAccessedAt),
			attribute.Stringer("args[3].expiresAt", expiresAt),
		),
	)
	defer span.End()

	return database.RunTx(spanCtx, r.txMgr, func(ctx context.Context, tx *gorm.DB) error {
		err := tx.Exec(
			`UPDATE sessions SET last_accessed_at = $1, expires_at = $2 WHERE id = $3`,
			lastAccessedAt,
			expiresAt,
			sessionID,
		).Error

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return err
		}

		span.SetStatus(codes.Ok, "")

		return nil
	})
}

func (r *PostgresRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	data map[string]any,
	expiresAt time.Time,
) (*snowdrop.SessionModel, error) {
	spanCtx, span := r.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(r.structName, "CreateSession"),
		trace.WithAttributes(
			attribute.Stringer("args[1].userID", userID),
			attribute.String("args[2].data", fmt.Sprintf("%v", data)),
			attribute.Stringer("args[3].expiresAt", expiresAt),
		),
	)
	defer span.End()

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return database.RunTxWithData(
		spanCtx,
		r.txMgr,
		func(ctx context.Context, tx *gorm.DB) (*snowdrop.SessionModel, error) {
			model := snowdrop.SessionModel{
				ID: sessionID,
				UserID: uuid.NullUUID{
					UUID:  lo.Ternary(userID == uuid.Nil, uuid.Nil, userID),
					Valid: userID != uuid.Nil,
				},
				CreatedAt:      now,
				LastAccessedAt: now,
				ExpiresAt:      expiresAt,
				// Data:           serializedData,
			}

			// var serializedData *[]byte

			// if len(data) > 0 {
			// 	dataAsBytes, marshalErr := json.Marshal(data)
			// 	if marshalErr != nil {
			// 		return nil, marshalErr
			// 	}

			// 	serializedData = &dataAsBytes
			// }

			if err != nil {
				return nil, err
			}

			result := tx.Clauses(clause.Returning{}).Create(&model)

			if result.Error != nil {
				return nil, result.Error
			}

			return &model, nil
		},
	)
}

func (r *PostgresRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return database.RunTx(ctx, r.txMgr, func(ctx context.Context, tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM sessions WHERE id = $1", sessionID).Error

		return err
	})
}

func (r *PostgresRepository) DeleteExpiredSessions(ctx context.Context) error {
	return database.RunTx(ctx, r.txMgr, func(ctx context.Context, tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM sessions WHERE expires_at < NOW()").Error

		return err
	})
}
