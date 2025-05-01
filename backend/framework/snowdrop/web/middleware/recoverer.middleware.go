package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/response"
)

func NewRecovererMiddleware(
	logger *slog.Logger,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := snowdrop.MustGetTracer(r)

			spanCtx, span := tracer.Start(r.Context(), "RecovererMiddleware")
			defer span.End(trace.WithStackTrace(true))

			req := r.WithContext(spanCtx)

			defer func() {
				rvr := recover()

				switch rvrValue := rvr.(type) {
				case nil:
					return
				case error:
					span.RecordError(rvrValue)
					span.SetStatus(codes.Error, rvrValue.Error())
				default:
					span.SetStatus(codes.Error, fmt.Sprintf("%v", rvr))
				}

				//nolint:err113,errorlint // rvr can be any value, not just an error
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				logger.ErrorContext(
					req.Context(),
					fmt.Sprintf("Panic recovered: %v", rvr),
					log.StacktraceLogAttr(fmt.Sprintf("%#v\n\n%v", rvr, string(debug.Stack()))),
				)

				if req.Header.Get("Connection") != "Upgrade" {
					response.NewBuilder(w, req).
						Status(http.StatusInternalServerError).
						JSON()
				}
			}()

			next.ServeHTTP(w, req)
		})
	}
}
