package httpsrv

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
	"internal.snowdrop/common/core"
)

const readHeaderTimeout = 2 * time.Second

func newHTTPServer(lc fx.Lifecycle, logger *slog.Logger) *http.Server {
	srv := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: readHeaderTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			logger.InfoContext(ctx, "Starting HTTP server at: "+srv.Addr)

			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					logger.ErrorContext(ctx, "HTTP server serve error: ", core.ErrorLogAttr(err))
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
