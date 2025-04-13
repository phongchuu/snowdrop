package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

func NewRecovererMiddleware(
	logger *slog.Logger,
) func(next http.Handler) http.Handler {
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

					logger.ErrorContext(ctx, "Critical error occurred: Panic recovered")

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
