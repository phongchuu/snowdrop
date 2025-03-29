package main

import (
	"embed"
	"flag"
	"log/slog"
	"net/http"
	"time"

	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"internal.snowdrop/common/core/config"
	"internal.snowdrop/common/core/database"
	"internal.snowdrop/common/core/log"
	"internal.snowdrop/common/core/session"
	"internal.snowdrop/common/core/trans"
	"internal.snowdrop/common/core/web"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/healthz"
	"internal.snowdrop/common/features/openapi"
	"internal.snowdrop/common/features/staticfile"
	"internal.snowdrop/common/features/usermgt"
)

var (
	//go:embed resources
	embedResourcesFolder embed.FS

	// AppVersion, AppRevision, and AppReleaseDate are global variables used to store application metadata.
	// These values are injected at build time using linker flags (e.g., -ldflags "-X main.AppVersion=1.0.0").
	// This is a valid use case for global variables as they are read-only and provide essential information
	// about the application build.
	//
	// Example linker flags:
	//   go build -ldflags "-X main.AppVersion=1.0.0 -X main.AppRevision=abc123 -X main.AppReleaseDate=2025-03-29"

	//nolint:gochecknoglobals // Globals used for build-time metadata
	AppVersion string

	//nolint:gochecknoglobals // Globals used for build-time metadata
	AppRevision string

	//nolint:gochecknoglobals // Globals used for build-time metadata
	AppReleaseDate string
)

func main() {
	time.Local = time.UTC

	isDebug := flag.Bool("debug", false, "Enable this flag to see debug log")
	flag.Parse()

	fx.New(
		fx.Supply(
			log.NewStdoutLogger(
				log.WithLogLevel(lo.Ternary(*isDebug, slog.LevelDebug, slog.LevelInfo)),
			),
		),
		config.NewModule(*isDebug, embedResourcesFolder),
		database.NewModule(),
		healthz.NewModule(),
		auth.NewModule(),
		web.NewModule(),
		openapi.NewModule(),
		session.NewModule(),
		staticfile.NewModule(),
		trans.NewModule(),
		usermgt.NewModule(),
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
