package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
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
	snowdropMiddleware "internal.snowdrop/framework/web/middleware"
)

type RouteParams struct {
	fx.In
	Config                     snowdrop.ConfigManager
	Logger                     *slog.Logger
	I18nBundle                 *i18n.Bundle
	HTTPRoutes                 []snowdrop.HTTPHandler `group:"http_routes"`
	SessionManager             snowdrop.SessionManager
	Tracer                     otelTrace.Tracer
	UniversalTranslator        *ut.UniversalTranslator
	GetPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn
	NotFoundHandler            NotFoundHandler         `optional:"true"`
	MethodNotAllowedHandler    MethodNotAllowedHandler `optional:"true"`
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

func defaultMiddlewares(p RouteParams) map[int]func(http.Handler) http.Handler {
	return map[int]func(http.Handler) http.Handler{
		0:  snowdropMiddleware.NewTraceMiddleware(p.Tracer),
		10: chiMiddleware.CleanPath,
		20: chiMiddleware.Compress(5),
		30: snowdropMiddleware.NewRequestSizeMiddleware(p.Config.GetMaxRequestSize()),
		40: snowdropMiddleware.NewI18nMiddleware(
			p.I18nBundle,
			p.GetPreferredUserLanguageFn,
		),
		50: snowdropMiddleware.NewUniversalTranslatorMiddleware(
			p.UniversalTranslator,
			p.GetPreferredUserLanguageFn,
		),
		60: snowdropMiddleware.NewRecovererMiddleware(p.Logger),
		70: chiMiddleware.Timeout(3 * time.Minute),
		80: snowdropMiddleware.NewSessionMiddleware(
			p.SessionManager,
			[]string{
				"GET /api/openapi",
				"GET /healthz",
				"GET /favicon.ico",
				"GET /public/*",
			},
		),
	}
}

func registerRoutes(r chi.Router, routes []snowdrop.HTTPHandler) {
	for _, route := range routes {
		routeTag := fmt.Sprintf("%s %s", route.Method(), route.Path())
		r.Method(route.Method(), route.Path(), otelhttp.WithRouteTag(routeTag, route))
	}
}

func useDefaultMiddlewares(r chi.Router, p RouteParams) {
	middlewares := defaultMiddlewares(p)
	middlewarePriorities := slices.Sorted(maps.Keys(middlewares))

	for _, priority := range middlewarePriorities {
		r.Use(middlewares[priority])
	}
}

// NewRouter initializes and returns a new HTTP router instance.
func NewRouter(params RouteParams) *chi.Mux {
	r := chi.NewRouter()

	routesByTag := map[snowdrop.RouteTag][]snowdrop.HTTPHandler{
		PublicRoute:  {},
		PrivateRoute: {},
	}

	for _, route := range params.HTTPRoutes {
		for _, tag := range lo.FindUniques(route.Tags()) {
			if _, exists := routesByTag[tag]; exists {
				routesByTag[tag] = append(routesByTag[tag], route)
			} else {
				params.Logger.Error(fmt.Sprintf(`the tag "%v" is not supported`, tag))
			}
		}
	}

	// Default routes
	r.Group(func(r chi.Router) {
		useDefaultMiddlewares(r, params)

		if params.NotFoundHandler != nil {
			r.NotFound(http.HandlerFunc(params.NotFoundHandler))
		}

		if params.MethodNotAllowedHandler != nil {
			r.MethodNotAllowed(http.HandlerFunc(params.MethodNotAllowedHandler))
		}
	})

	// Public routes
	r.Group(func(r chi.Router) {
		useDefaultMiddlewares(r, params)
		registerRoutes(r, routesByTag[PublicRoute])
	})

	// Private routes
	r.Group(func(r chi.Router) {
		useDefaultMiddlewares(r, params)
		r.Use(snowdropMiddleware.NewAuthMiddleware())
		registerRoutes(r, routesByTag[PrivateRoute])
	})

	_ = chi.Walk(
		r,
		func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
			params.Logger.Info(fmt.Sprintf("Registered route: [%s] %s", method, route))

			return nil
		},
	)

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
