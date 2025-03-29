package database

import (
	"context"
	"database/sql"
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

// newDatabase creates a new SQL database connection using the URL from the configuration manager. It parses the database URL, configures a connection pool with tracing enabled, and registers lifecycle hooks to ping the connection on startup and close it on shutdown.
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
// GenerateUUIDv7PgType generates a new version 7 UUID and returns it as a pgtype.UUID. The returned UUID is marked as valid and the function panics if UUID generation fails.
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
