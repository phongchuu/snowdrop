package staticfile

import (
	"go.uber.org/fx"
	"internal.snowdrop/framework/web"
)

func NewModule() fx.Option {
	return fx.Module(
		"StaticFileModule",
		fx.Provide(
			web.HTTPRoute(NewStaticRoute),
		),
	)
}
