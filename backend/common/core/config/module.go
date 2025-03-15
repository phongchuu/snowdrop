package config

import (
	"io/fs"

	"go.uber.org/fx"
)

// NewModule creates a new Fx module for application configuration.
// It takes an embedded resources folder as an argument and provides an AppConfig instance.
//
// Parameters:
//   - embedResourcesFolder: an fs.FS instance representing the embedded resources folder.
//
// Returns:
//   - fx.Option: an Fx module option that provides the AppConfig instance.
func NewModule(embedResourcesFolder fs.FS) fx.Option {
	return fx.Module(
		"ConfigModule",
		fx.Provide(func() (AppConfig, error) {
			return NewAppConfig(embedResourcesFolder)
		}),
	)
}
