package healthz

import (
	"net/http"

	"internal.snowdrop/common/core"
)

type HealthCheckRoute struct{}

var _ core.HTTPHandler = (*HealthCheckRoute)(nil)

func NewHealthCheckRoute() *HealthCheckRoute {
	return &HealthCheckRoute{}
}

func (h *HealthCheckRoute) Method() string {
	return http.MethodGet
}

func (h *HealthCheckRoute) Path() string {
	return "/healthz"
}

func (h *HealthCheckRoute) Tags() []core.RouteTag {
	return []core.RouteTag{core.PublicRoute}
}

func (h *HealthCheckRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resCtrl := core.NewResponseController(w, r)
	resCtrl.Status(http.StatusNoContent)
}
