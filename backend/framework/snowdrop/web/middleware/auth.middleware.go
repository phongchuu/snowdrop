package middleware

import (
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/response"
)

func NewAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := snowdrop.MustGetTracer(r)

			spanCtx, span := tracer.Start(r.Context(), "AuthMiddleware")
			defer span.End()

			newReq := r.WithContext(spanCtx)
			session, err := snowdrop.GetCurrentSession(newReq)

			if err != nil && !errors.Is(err, snowdrop.ErrNoSession) {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				response.NewBuilder(w, newReq).
					Status(http.StatusInternalServerError).
					JSON()

				return
			}

			if errors.Is(err, snowdrop.ErrNoSession) || !session.UserID.Valid {
				span.SetStatus(codes.Error, "Unauthorized")
				response.NewBuilder(w, newReq).
					Status(http.StatusUnauthorized).
					JSON()

				return
			}

			next.ServeHTTP(w, newReq)
		})
	}
}
