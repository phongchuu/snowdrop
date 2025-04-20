package opentelemetry

import "go.uber.org/fx"

func NewModule() fx.Option {
	return fx.Module(
		"OpentelemetryModule",
		fx.Provide(
			SetupOtelSDK,
		),
	)
}
