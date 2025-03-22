package config

import (
	"io/fs"

	"go.uber.org/fx"
)

type Manager interface {
	IsDebugMode() bool

	// GetEmbedResourceFolder returns the embedded resource folder as a filesystem (fs.FS).
	// This folder contains the embedded resources used by the application.
	GetEmbedResourceFolder() fs.FS

	// GetDatabaseURL constructs and returns the database connection URL
	// using the configuration values stored in the AppConfig instance.
	// It retrieves the database username, password, host, port, and name
	// from the viper instance and formats them into a PostgreSQL connection string.
	//
	// Returns: The formatted PostgreSQL connection URL.
	GetDatabaseURL() string
}

// NewModule creates a new Fx module for application configuration.
// It takes an embedded resources folder as an argument and provides an AppConfig instance.
//
// Parameters:
//   - embedResourcesFolder: an fs.FS instance representing the embedded resources folder.
//
// Returns:
//   - fx.Option: an Fx module option that provides the AppConfig instance.
func NewModule(debugMode bool, embedResourcesFolder fs.FS) fx.Option {
	return fx.Module(
		"ConfigModule",
		fx.Provide(
			fx.Annotate(func() (AppConfig, error) {
				return NewAppConfig(AppConfigParams{
					IsDebugMode:          debugMode,
					EmbedResourcesFolder: embedResourcesFolder,
				})
			}, fx.As(new(Manager))),
		),
	)
}
