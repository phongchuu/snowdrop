package main

import (
	"embed"

	"go.uber.org/fx"
	"internal.snowdrop/common/core"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/healthz"
	"internal.snowdrop/common/features/openapi"
	"internal.snowdrop/common/features/staticfile"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework/command"
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
)

func main() {
	snowdrop.Start([]fx.Option{
		core.NewModule(core.ModuleConfig{
			Version:              AppVersion,
			Revision:             AppRevision,
			EmbedResourcesFolder: embedResourcesFolder,
		}),
		healthz.NewModule(),
		auth.NewModule(),
		openapi.NewModule(),
		staticfile.NewModule(),
		usermgt.NewModule(),
	})
}
