package core

import (
	"io/fs"

	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

type ModuleConfig struct {
	Version              string
	Revision             string
	EmbedResourcesFolder fs.FS
}

func NewModule(fwCfg ModuleConfig) fx.Option {
	return fx.Module(
		"CoreModule",
		fx.Provide(
			fx.Annotate(func() (AppConfig, error) {
				return NewAppConfig(fwCfg)
			}, fx.As(new(snowdrop.ConfigManager))),
			NewGetPreferredUserLanguageFn,
		),
	)
}
