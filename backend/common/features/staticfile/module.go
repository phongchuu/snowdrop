package staticfile

import (
	"io/fs"

	"go.uber.org/fx"
	"internal.snowdrop/common/core/web"
)

func NewModule(embedResourcesFolder fs.FS) fx.Option {
	return fx.Module(
		"StaticModule",
		fx.Provide(
			web.HTTPRoute(func() *FileServerRoute {
				return NewStaticRoute(embedResourcesFolder)
			}),
		),
	)
}
