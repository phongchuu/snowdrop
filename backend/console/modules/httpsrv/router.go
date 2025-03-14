package httpsrv

import (
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"internal.snowdrop/common/core"
)

type RouteParams struct {
	fx.In
	HTTPRoutes []core.HTTPHandler `group:"http_routes"`
}

// NewRouter initializes and returns a new HTTP router instance.
func NewRouter(params RouteParams) *chi.Mux {
	r := chi.NewRouter()

	publicRoutes := []core.HTTPHandler{}
	privateRoutes := []core.HTTPHandler{}

	for _, route := range params.HTTPRoutes {
		if slices.Contains(route.Tags(), core.PrivateRoute) {
			privateRoutes = append(privateRoutes, route)
		} else {
			publicRoutes = append(publicRoutes, route)
		}
	}

	middlewares := map[int]func(http.Handler) http.Handler{
		0:  middleware.CleanPath,
		60: middleware.Timeout(time.Minute),
	}

	// Public routes
	r.Group(func(r chi.Router) {
		middlewarePriorities := lo.Keys(middlewares)
		slices.Sort(middlewarePriorities)
		r.Use(lo.Values(middlewares)...)

		for _, route := range publicRoutes {
			r.Method(route.Method(), route.Path(), route)
		}
	})

	// Private routes
	r.Group(func(r chi.Router) {
		middlewarePriorities := lo.Keys(middlewares)
		slices.Sort(middlewarePriorities)
		r.Use(lo.Values(middlewares)...)

		for _, route := range privateRoutes {
			r.Method(route.Method(), route.Path(), route)
		}
	})

	return r
}
