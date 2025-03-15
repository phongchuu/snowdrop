package web

import (
	"net/http"

	"go.uber.org/fx"
)

func NewModule() fx.Option {
	return fx.Module(
		"WebModule",
		fx.Provide(
			fx.Annotate(NewRouter, fx.As(new(http.Handler))),
			newHTTPServer,
			NewSchemaDecoder,
		),
	)
}
