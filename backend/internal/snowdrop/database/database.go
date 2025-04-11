package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	log "github.com/mcosta74/pgx-slog"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

type gooseSlogger struct {
	Logger *slog.Logger
}

var _ goose.Logger = (*gooseSlogger)(nil)

func (s gooseSlogger) Printf(format string, v ...any) {
	s.Logger.Info(fmt.Sprintf(format, v...))
}

func (s gooseSlogger) Fatalf(format string, v ...any) {
	s.Logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

// newDatabase creates a new database connection pool.
// Registers lifecycle hooks for connection health checks and cleanup.
func newDatabase(
	lc fx.Lifecycle,
	logger *slog.Logger,
	config snowdrop.ConfigManager,
) (*sql.DB, error) {
	cfg, _ := pgxpool.ParseConfig(config.GetDatabaseURL())
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   log.NewLogger(logger),
		LogLevel: tracelog.LogLevelInfo,
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
			goose.SetLogger(&gooseSlogger{Logger: logger})
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
