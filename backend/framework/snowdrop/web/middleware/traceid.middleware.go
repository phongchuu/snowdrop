package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func NewTraceMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			props := otel.GetTextMapPropagator()
			props.Inject(r.Context(), propagation.HeaderCarrier(w.Header()))
			next.ServeHTTP(w, r)
		})
	}
}
