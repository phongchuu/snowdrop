package httpsrv

import (
	"net/http"

	"go.uber.org/fx"
)

func NewHTTPServerModule() fx.Option {
	return fx.Module(
		"HttpServerModule",
		fx.Provide(
			fx.Annotate(NewRouter, fx.As(new(http.Handler))),
			newHTTPServer,
		),
	)
}
