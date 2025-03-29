package web

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"internal.snowdrop/common/core/log"
	"internal.snowdrop/common/core/trans"
)

type RouteTag = int

type HTTPHandler interface {
	http.Handler

	Method() string
	Path() string
	Tags() []RouteTag
}

type RouteParams struct {
	fx.In
	I18nBundle *i18n.Bundle
	HTTPRoutes []HTTPHandler `group:"http_routes"`
}

// HTTPRoute is a helper function that annotates a given function to be used as an HTTP handler
// within the Fx framework. It registers the function as a new HTTPHandler and tags the result
// with the group "http_routes" for dependency injection.
//
// Parameters:
//   - function: The function to be annotated.
//
// Returns:
//   - Annotated function ready to be used as an HTTP handler in the Fx framework.
func HTTPRoute(function any) any {
	return fx.Annotate(
		function,
		fx.As(new(HTTPHandler)),
		fx.ResultTags(`group:"http_routes"`),
	)
}

// NewRouter initializes and returns a new HTTP router instance.
func NewRouter(params RouteParams) *chi.Mux {
	r := chi.NewRouter()

	publicRoutes := []HTTPHandler{}
	privateRoutes := []HTTPHandler{}

	for _, route := range params.HTTPRoutes {
		if slices.Contains(route.Tags(), PrivateRoute) {
			privateRoutes = append(privateRoutes, route)
		} else {
			publicRoutes = append(publicRoutes, route)
		}
	}

	middlewares := map[int]func(http.Handler) http.Handler{
		0:  middleware.CleanPath,
		30: trans.WithI18nMiddleware(params.I18nBundle),
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

func newHTTPServer(
	lc fx.Lifecycle,
	logger *slog.Logger,
	httpHandler http.Handler,
) *http.Server {
	srv := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: readHeaderTimeout,
		Handler:           httpHandler,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			logger.InfoContext(ctx, "Starting HTTP server at: "+srv.Addr)

			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.ErrorContext(ctx, "HTTP server serve error: ", log.ErrorLogAttr(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
