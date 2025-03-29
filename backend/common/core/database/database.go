package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	log "github.com/mcosta74/pgx-slog"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"internal.snowdrop/common/core/config"
)

type Database interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func newDatabase(
	lc fx.Lifecycle,
	logger *slog.Logger,
	config config.Manager,
) (*sql.DB, error) {
	cfg, _ := pgxpool.ParseConfig(config.GetDatabaseURL())
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   log.NewLogger(logger),
		LogLevel: tracelog.LogLevelDebug,
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.PingContext(ctx)
		},
		OnStop: func(_ context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}

// executeDatabaseUpgrade sets up a database upgrade process using the provided
// lifecycle, database connection, and application configuration. It registers
// a start hook that initializes the goose migration tool with the embedded
// resource folder and the PostgreSQL dialect, then runs the migrations located
// in the "resources/migrations" directory.
//
// Parameters:
//   - lc: The lifecycle to which the start hook will be appended.
//   - db: The bun.DB instance representing the database connection.
//   - config: The application configuration containing the embedded resource folder.
//
// Returns:
//   - An error if the goose migration setup or execution fails, otherwise nil.
func executeDatabaseUpgrade(
	lc fx.Lifecycle,
	logger *slog.Logger,
	db *sql.DB,
	config config.Manager,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			goose.SetBaseFS(config.GetEmbedResourceFolder())
			goose.SetLogger(&gooseSlogger{Logger: logger})
			goose.SetTableName("public.migration_history")

			return goose.UpContext(ctx, db, "resources/migrations")
		},
	})
}

// GenerateUUIDv7PgType generates a new UUID version 7 and returns it as a pgtype.UUID.
// If there is an error generating the UUID, the function will panic.
// The returned pgtype.UUID will have the generated UUID bytes and will be marked as valid.
func GenerateUUIDv7PgType() pgtype.UUID {
	uuid, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}
}

func configureTransaction(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "SELECT set_config('session.requester', $1, true)", "xxx")

	return err
}

func GetTransaction(ctx context.Context, db Database) (*sql.Tx, bool, error) {
	if tx, ok := db.(*sql.Tx); ok {
		if err := configureTransaction(ctx, tx); err != nil {
			_ = tx.Rollback()

			return nil, false, err
		}

		return tx, false, nil
	}

	if pool, ok := db.(*sql.DB); ok {
		tx, err := pool.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return nil, false, err
		}

		if err := configureTransaction(ctx, tx); err != nil {
			_ = tx.Rollback()

			return nil, false, err
		}

		return tx, true, nil
	}

	return nil, false, errors.ErrUnsupported
}

func AutoRollbackOrCommit(tx *sql.Tx, err error) error {
	if err != nil {
		return tx.Rollback()
	}

	if err := tx.Commit(); err != nil {
		return tx.Rollback()
	}

	return nil
}
