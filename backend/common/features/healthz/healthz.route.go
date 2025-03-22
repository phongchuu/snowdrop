package healthz

import (
	"net/http"

	"internal.snowdrop/common/core/web"
)

type HealthCheckRoute struct{}

var _ web.HTTPHandler = (*HealthCheckRoute)(nil)

func NewHealthCheckRoute() HealthCheckRoute {
	return HealthCheckRoute{}
}

// Method implements web.HTTPHandler.
func (h HealthCheckRoute) Method() string {
	return http.MethodGet
}

// Path implements web.HTTPHandler.
func (h HealthCheckRoute) Path() string {
	return "/healthz"
}

// Tags implements web.HTTPHandler.
func (h HealthCheckRoute) Tags() []web.RouteTag {
	return []web.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (h HealthCheckRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.NewResponseBuilder(w, r).NoContent()
}
