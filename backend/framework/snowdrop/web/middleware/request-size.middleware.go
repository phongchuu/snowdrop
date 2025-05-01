package middleware

import (
	"net/http"

	snowdrop "internal.snowdrop/framework"
)

func NewRequestSizeMiddleware(size int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := snowdrop.MustGetTracer(r)

			spanCtx, span := tracer.Start(r.Context(), "RequestSizeMiddleware")
			defer span.End()

			req := r.WithContext(spanCtx)
			r2 := *req
			r2.Body = http.MaxBytesReader(w, req.Body, size)
			next.ServeHTTP(w, &r2)
		})
	}
}
