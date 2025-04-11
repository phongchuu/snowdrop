package healthz

import (
	"go.uber.org/fx"
	"internal.snowdrop/framework/web"
)

func NewModule() fx.Option {
	return fx.Module(
		"HealthzModule",
		fx.Provide(web.HTTPRoute(NewHealthCheckRoute)),
	)
}
