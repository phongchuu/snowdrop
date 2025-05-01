package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	snowdrop "internal.snowdrop/framework"
)

func NewTraceMiddleware(tracer trace.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spanCtx, span := tracer.Start(r.Context(), "TraceMiddleware")
			defer span.End()

			req := r.WithContext(spanCtx)
			props := otel.GetTextMapPropagator()
			props.Inject(req.Context(), propagation.HeaderCarrier(w.Header()))
			next.ServeHTTP(w, snowdrop.WithTracer(req, tracer))
		})
	}
}
