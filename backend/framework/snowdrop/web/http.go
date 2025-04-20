package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	ut "github.com/go-playground/universal-translator"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/session"
	snowdropMiddleware "internal.snowdrop/framework/web/middleware"
)

type RouteParams struct {
	fx.In
	Config                     snowdrop.ConfigManager
	Logger                     *slog.Logger
	I18nBundle                 *i18n.Bundle
	HTTPRoutes                 []snowdrop.HTTPHandler `group:"http_routes"`
	SessionManager             *session.Manager
	Tracer                     otelTrace.Tracer
	UniversalTranslator        *ut.UniversalTranslator
	GetPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn
}

// HTTPRoute annotates a function as an HTTP handler for Fx DI.
// Tags it for the "http_routes" group.
func HTTPRoute(handlerFn any) any {
	return fx.Annotate(
		handlerFn,
		fx.As(new(snowdrop.HTTPHandler)),
		fx.ResultTags(`group:"http_routes"`),
	)
}

// NewRouter initializes and returns a new HTTP router instance.
func NewRouter(params RouteParams) *chi.Mux {
	r := chi.NewRouter()

	publicRoutes := []snowdrop.HTTPHandler{}
	privateRoutes := []snowdrop.HTTPHandler{}

	for _, route := range params.HTTPRoutes {
		if slices.Contains(route.Tags(), PrivateRoute) {
			privateRoutes = append(privateRoutes, route)
		} else {
			publicRoutes = append(publicRoutes, route)
		}
	}

	middlewares := map[int]func(http.Handler) http.Handler{
		0: snowdropMiddleware.NewTraceMiddleware(),
		1: chiMiddleware.CleanPath,
		2: chiMiddleware.Compress(5),
		3: snowdropMiddleware.NewRequestSizeMiddleware(params.Config.GetMaxRequestSize()),
		60: snowdropMiddleware.NewI18nMiddleware(
			params.I18nBundle,
			params.GetPreferredUserLanguageFn,
		),
		61: snowdropMiddleware.NewRecovererMiddleware(params.Logger),
		80: snowdropMiddleware.NewUniversalTranslatorMiddleware(
			params.UniversalTranslator,
			params.GetPreferredUserLanguageFn,
		),
		100: params.SessionManager.Middleware,
		120: chiMiddleware.Timeout(3 * time.Minute),
	}

	// Public routes
	r.Group(func(r chi.Router) {
		middlewarePriorities := lo.Keys(middlewares)
		slices.Sort(middlewarePriorities)
		r.Use(lo.Values(middlewares)...)

		for _, route := range publicRoutes {
			routeTag := fmt.Sprintf("%s %s", route.Method(), route.Path())
			r.Method(route.Method(), route.Path(), otelhttp.WithRouteTag(routeTag, route))
		}
	})

	// Private routes
	r.Group(func(r chi.Router) {
		middlewarePriorities := lo.Keys(middlewares)
		slices.Sort(middlewarePriorities)
		r.Use(lo.Values(middlewares)...)

		for _, route := range privateRoutes {
			routeTag := fmt.Sprintf("%s %s", route.Method(), route.Path())
			r.Method(route.Method(), route.Path(), otelhttp.WithRouteTag(routeTag, route))
		}
	})

	return r
}

// newHTTPServer creates a configured HTTP server with OpenTelemetry instrumentation.
func newHTTPServer(
	lc fx.Lifecycle,
	logger *slog.Logger,
	httpHandler http.Handler,
	config snowdrop.ConfigManager,
) *http.Server {
	srv := &http.Server{
		Addr:              net.JoinHostPort("", config.GetAppPort()),
		ReadHeaderTimeout: readHeaderTimeout,
		Handler: otelhttp.NewHandler(httpHandler,
			"",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}),
		),
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
