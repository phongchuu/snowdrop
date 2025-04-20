package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"internal.snowdrop/framework/response"
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

					logger.ErrorContext(
						ctx,
						fmt.Sprintf("Panic recovered: %v", rvr),
						slog.String("stacktrace", fmt.Sprintf("%#v\n\n%s", rvr, debug.Stack())),
					)

					if r.Header.Get("Connection") != "Upgrade" {
						response.NewBuilder(w, r).
							Status(http.StatusInternalServerError).
							JSON()
					}
				}
			}(r.Context())

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
