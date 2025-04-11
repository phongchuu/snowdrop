package openapi

import (
	"go.uber.org/fx"
	"internal.snowdrop/framework/web"
)

func NewModule() fx.Option {
	return fx.Module(
		"StaticModule",
		fx.Provide(
			web.HTTPRoute(NewDocumentationRoute),
		),
	)
}
