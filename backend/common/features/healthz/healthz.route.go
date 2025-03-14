package healthz

import (
	"net/http"

	"internal.snowdrop/common/core/web"
)

type HealthCheckRoute struct{}

var _ web.HTTPHandler = (*HealthCheckRoute)(nil)

func NewHealthCheckRoute() *HealthCheckRoute {
	return &HealthCheckRoute{}
}

func (h *HealthCheckRoute) Method() string {
	return http.MethodGet
}

func (h *HealthCheckRoute) Path() string {
	return "/healthz"
}

func (h *HealthCheckRoute) Tags() []web.RouteTag {
	return []web.RouteTag{web.PublicRoute}
}

func (h *HealthCheckRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resCtrl := web.NewResponseController(w, r)
	resCtrl.Status(http.StatusNoContent)
}
