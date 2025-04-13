package web

import (
	"net/http"

	"go.uber.org/fx"
	"internal.snowdrop/framework/log"
)

func NewModule() fx.Option {
	return fx.Module(
		"WebModule",
		fx.Provide(
			fx.Annotate(NewRouter, fx.As(new(http.Handler))),
			newHTTPServer,
			log.SetupOtelSDK,
		),
	)
}
