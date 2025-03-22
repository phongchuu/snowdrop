package auth

import (
	"go.uber.org/fx"
	"internal.snowdrop/common/core/web"
)

func NewModule() fx.Option {
	return fx.Module(
		"AuthModule",
		fx.Provide(web.HTTPRoute(NewRegisterRoute)),
	)
}
