package database

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

// newDatabase creates a new database connection pool.
// Registers lifecycle hooks for connection health checks and cleanup.
func newDatabase(
	lc fx.Lifecycle,
	logger *slog.Logger,
	config snowdrop.ConfigManager,
) (*sql.DB, error) {
	cfg, _ := pgxpool.ParseConfig(config.GetDatabaseURL())
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   NewPgxLogger(logger),
		LogLevel: tracelog.LogLevelTrace,
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

// executeDatabaseUpgrade runs database migrations on application startup.
// Uses goose migration tool with configurations from the provided ConfigManager.
func executeDatabaseUpgrade(
	lc fx.Lifecycle,
	logger *slog.Logger,
	db *sql.DB,
	config snowdrop.ConfigManager,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			goose.SetBaseFS(config.GetEmbedResourceFolder())
			goose.SetLogger(&gooseSlogger{logger: logger})
			goose.SetTableName("public.migration_history")

			return goose.UpContext(ctx, db, "resources/migrations")
		},
	})
}

func NewModule() fx.Option {
	return fx.Module(
		"DatabaseModule",
		fx.Provide(
			newDatabase,
			fx.Annotate(
				NewTransactionManager,
				fx.As(new(snowdrop.TransactionManager)),
			),
		),
		fx.Invoke(executeDatabaseUpgrade),
	)
}
