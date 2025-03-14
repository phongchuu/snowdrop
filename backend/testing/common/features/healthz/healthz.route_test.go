package healthz_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"internal.snowdrop/common/core/web"
	"internal.snowdrop/common/features/healthz"
	"internal.snowdrop/console/modules/httpsrv"
)

func TestHealthCheckRoute(t *testing.T) {
	router := httpsrv.NewRouter(httpsrv.RouteParams{
		HTTPRoutes: []web.HTTPHandler{healthz.NewHealthCheckRoute()},
	})

	assert.True(t, router.Match(chi.NewRouteContext(), http.MethodGet, "/healthz"))
}

func TestHealthCheckRouteServeHTTP(t *testing.T) {
	healthzRoute := healthz.NewHealthCheckRoute()
	router := httpsrv.NewRouter(httpsrv.RouteParams{
		HTTPRoutes: []web.HTTPHandler{healthzRoute},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
