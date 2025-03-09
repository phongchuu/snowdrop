package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/fx"
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

// newDatabase initializes a new Bun database connection using the provided
// lifecycle and application configuration. It sets up the connection pool
// and registers lifecycle hooks to manage the database connection's lifecycle.
//
// Parameters:
//   - lc: The Fx lifecycle to which the database connection hooks will be appended.
//   - config: The application configuration containing the database URL.
//
// Returns:
//   - *bun.DB: A pointer to the initialized Bun database instance.
//   - error: An error if the database connection could not be established.
func newDatabase(
	lc fx.Lifecycle,
	config AppConfig,
) (*bun.DB, error) {
	pool, err := pgxpool.New(context.Background(), config.GetDatabaseURL())
	if err != nil {
		return nil, err
	}

	sqldb := stdlib.OpenDBFromPool(pool)
	db := bun.NewDB(sqldb, pgdialect.New())

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
func executeDatabaseUpgrade(lc fx.Lifecycle, logger *slog.Logger, db *bun.DB, config AppConfig) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			goose.SetBaseFS(config.GetEmbedResourceFolder())
			goose.SetLogger(&gooseSlogger{Logger: logger})
			goose.SetTableName("public.migration_history")

			if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
				return err
			}

			return goose.UpContext(ctx, db.DB, "resources/migrations")
		},
	})
}

// NewDatabaseModule returns an fx.Option that provides the database module for the application.
// It includes the necessary dependencies and invokes the database upgrade execution.
//
// The module includes:
// - fx.Provide(newDatabase): Provides the database instance.
// - fx.Invoke(executeDatabaseUpgrade): Executes the database upgrade process.
func NewDatabaseModule() fx.Option {
	return fx.Module(
		"DatabaseModule",
		fx.Provide(newDatabase),
		fx.Invoke(executeDatabaseUpgrade),
	)
}
