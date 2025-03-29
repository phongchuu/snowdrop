package database

import (
	"go.uber.org/fx"
)

// NewModule returns an fx.Option that provides the database module for the application.
// It includes the necessary dependencies and invokes the database upgrade execution.
//
// The module includes:
// - fx.Provide(newDatabase): Provides the database instance.
// NewModule creates an fx.Option that sets up the database module for the application.
// It provides a new database instance via newDatabase and an annotated TransactionManager,
// ensuring that the TransactionManager is correctly recognized by the dependency injection system.
// The module also invokes executeDatabaseUpgrade to perform any required database upgrades during initialization.
func NewModule() fx.Option {
	return fx.Module(
		"DatabaseModule",
		fx.Provide(
			newDatabase,
			fx.Annotate(
				NewTransactionManager,
				fx.As(new(TransactionManager)),
			),
		),
		fx.Invoke(executeDatabaseUpgrade),
	)
}
