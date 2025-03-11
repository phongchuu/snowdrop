package healthz

import (
	"go.uber.org/fx"
	"internal.snowdrop/common/core"
)

func NewHealthzModule() fx.Option {
	return fx.Module(
		"HealthzModule",
		fx.Provide(core.Route(NewHealthCheckRoute)),
	)
}
