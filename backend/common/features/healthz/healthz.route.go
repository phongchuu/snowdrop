package healthz

import (
	"net/http"

	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/web"
)

type HealthCheckRoute struct{}

var _ snowdrop.HTTPHandler = (*HealthCheckRoute)(nil)

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
func (h HealthCheckRoute) Tags() []snowdrop.RouteTag {
	return []snowdrop.RouteTag{web.PublicRoute}
}

// ServeHTTP implements web.HTTPHandler.
func (h HealthCheckRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.NewResponseBuilder(w, r).NoContent()
}
