package healthz

import (
	"go.uber.org/fx"
	"internal.snowdrop/common/core/web"
)

func NewHealthzModule() fx.Option {
	return fx.Module(
		"HealthzModule",
		fx.Provide(web.HTTPRoute(NewHealthCheckRoute)),
	)
}
