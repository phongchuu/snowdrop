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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/trans"
)

type RecovererMiddleware func(next http.Handler) http.Handler

type RouteParams struct {
	fx.In
	I18nBundle                    *i18n.Bundle
	HTTPRoutes                    []snowdrop.HTTPHandler `group:"http_routes"`
	SessionManager                *session.Manager
	I18nMiddleware                trans.I18nMiddleware
	UniversalTranslatorMiddleware UniversalTranslatorMiddleware
	RecovererMiddleware           RecovererMiddleware
	Logger                        *slog.Logger
	Tracer                        otelTrace.Tracer
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
		0: otelhttp.NewMiddleware(
			"__placeholder__",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}),
		),
		1:   middleware.CleanPath,
		20:  params.RecovererMiddleware,
		60:  params.I18nMiddleware,
		80:  params.UniversalTranslatorMiddleware,
		100: params.SessionManager.Middleware,
		120: middleware.Timeout(time.Minute),
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

func NewRecovererMiddleware(
	logger *slog.Logger,
) RecovererMiddleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func(ctx context.Context) {
				if rvr := recover(); rvr != nil {
					//nolint:err113,errorlint // rvr can be any value, not just an error
					if rvr == http.ErrAbortHandler {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					logger.ErrorContext(
						ctx,
						"Critical error occurred: Panic recovered",
						// slog.Any("panic", rvr),
						// slog.Any("stack_trace", string(debug.Stack())),
					)

					if r.Header.Get("Connection") != "Upgrade" {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}(r.Context())

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
