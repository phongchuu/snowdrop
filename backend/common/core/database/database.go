package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/fx"
	"internal.snowdrop/common/core/config"
)

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
	config config.Manager,
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
func executeDatabaseUpgrade(lc fx.Lifecycle, logger *slog.Logger, db *bun.DB, config config.Manager) {
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
