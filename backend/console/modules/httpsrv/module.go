package httpsrv

import (
	"go.uber.org/fx"
)

func NewHTTPServerModule() fx.Option {
	return fx.Module(
		"HttpServerModule",
		fx.Provide(
			newRouter,
			newHTTPServer,
		),
	)
}
