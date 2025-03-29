package database

import (
	"go.uber.org/fx"
)

// NewModule returns an fx.Option that provides the database module for the application.
// It includes the necessary dependencies and invokes the database upgrade execution.
//
// The module includes:
// - fx.Provide(newDatabase): Provides the database instance.
// - fx.Invoke(executeDatabaseUpgrade): Executes the database upgrade process.
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
