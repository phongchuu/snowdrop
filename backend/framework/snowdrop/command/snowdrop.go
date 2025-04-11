package command

import (
	"net/http"
	"time"

	"go.uber.org/fx"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/trans"
	"internal.snowdrop/framework/web"
)

// Start initializes and starts the Snowdrop application with the provided options.
func Start(opts []fx.Option) {
	time.Local = time.UTC

	options := []fx.Option{
		// config.NewModule(frameworkConfig),
		database.NewModule(),
		web.NewModule(),
		session.NewModule(),
		trans.NewModule(),
		fx.Invoke(func(*http.Server) {}),
	}

	options = append(options, opts...)

	if err := fx.ValidateApp(options...); err != nil {
		panic(err)
	}

	fx.New(options...).Run()
}
