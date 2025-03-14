package main

import (
	"embed"
	"flag"
	"log/slog"
	"net/http"

	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"internal.snowdrop/common/core/config"
	"internal.snowdrop/common/core/database"
	"internal.snowdrop/common/core/log"
	"internal.snowdrop/common/features/healthz"
	"internal.snowdrop/console/modules/httpsrv"
)

var (
	//go:embed resources
	embedResourcesFolder embed.FS

	AppVersion     string
	AppRevision    string
	AppReleaseDate string
)

func main() {
	isDebug := flag.Bool("debug", false, "Enable this flag to see debug log")
	flag.Parse()

	fx.New(
		fx.Supply(log.NewStdoutLogger(log.WithLogLevel(lo.Ternary(*isDebug, slog.LevelDebug, slog.LevelInfo)))),
		config.NewConfigModule(embedResourcesFolder),
		database.NewDatabaseModule(),
		httpsrv.NewHTTPServerModule(),
		healthz.NewHealthzModule(),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			fxEventLogger := &fxevent.SlogLogger{
				Logger: logger,
			}
			fxEventLogger.UseLogLevel(slog.LevelDebug)
			return fxEventLogger
		}),
		fx.Invoke(func(logger *slog.Logger, _ *http.Server) {
			logger.Info("App information",
				slog.String("version", AppVersion),
				slog.String("revision", AppRevision),
				slog.String("release_date", AppReleaseDate),
			)
		}),
	).Run()
}
