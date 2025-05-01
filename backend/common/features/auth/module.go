package auth

import (
	"go.uber.org/fx"
	"internal.snowdrop/framework/web"
)

func NewModule() fx.Option {
	return fx.Module(
		"AuthModule",
		fx.Provide(
			web.HTTPRoute(NewRegisterRoute),
			web.HTTPRoute(NewLoginRoute),
			web.HTTPRoute(NewProfileRoute),
		),
	)
}
