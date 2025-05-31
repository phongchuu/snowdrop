package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	snowdrop "internal.snowdrop/framework"
)

type DatabaseResult struct {
	fx.Out
	GormDB *gorm.DB
	SQLDB  *sql.DB
}

func setExecutor(tx *gorm.DB) {
	currentSession, err := snowdrop.GetCurrentSessionFromCtx(tx.Statement.Context)
	if err != nil || errors.Is(err, snowdrop.ErrNoSession) {
		err = tx.Exec(
			`SELECT set_config('session.requester', u.id::VARCHAR(255), true)
            FROM public.users u
            WHERE u.username = $1;`,
			"system",
		).Error
	}

	if currentSession != nil {
		if !currentSession.UserID.Valid {
			err = tx.Exec(
				`SELECT set_config('session.requester', u.id::VARCHAR(255), true)
                FROM public.users u
                WHERE u.username = $1;`,
				"annonymous",
			).Error
		}

		if currentSession.UserID.Valid {
			err = tx.Exec(
				`SELECT set_config('session.requester', $1, true);`,
				currentSession.UserID.UUID.String(),
			).Error
		}
	}
}

// NewDatabase creates a new database connection pool.
// Registers lifecycle hooks for connection health checks and cleanup.
func NewDatabase(
	lc fx.Lifecycle,
	logger *slog.Logger,
	config snowdrop.ConfigManager,
) (DatabaseResult, error) {
	print("Initializing database connection...")
	print("Database URL: ", config.GetDatabaseURL())
	db, err := gorm.Open(postgres.Open(config.GetDatabaseURL()), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		return DatabaseResult{}, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return DatabaseResult{}, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db.Callback().Create().Before("gorm:before_create").Register("audit:executor", setExecutor)
	db.Callback().Update().Before("gorm:before_update").Register("audit:executor", setExecutor)
	db.Callback().Delete().Before("gorm:before_delete").Register("audit:executor", setExecutor)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(_ context.Context) error {
			return sqlDB.Close()
		},
	})

	return DatabaseResult{
		GormDB: db,
		SQLDB:  sqlDB,
	}, nil
}

// ExecuteDatabaseUpgrade runs database migrations on application startup.
// Uses goose migration tool with configurations from the provided ConfigManager.
func ExecuteDatabaseUpgrade(
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
			NewDatabase,
			fx.Annotate(
				NewTransactionManager,
				fx.As(new(snowdrop.TransactionManager)),
			),
		),
		fx.Invoke(ExecuteDatabaseUpgrade),
	)
}
