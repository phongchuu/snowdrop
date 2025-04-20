package command

import (
	"net/http"
	"time"

	"go.uber.org/fx"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/opentelemetry"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/validation"
	"internal.snowdrop/framework/web"
)

// Start initializes and starts the Snowdrop application with the provided options.
func Start(opts []fx.Option) {
	time.Local = time.UTC

	options := []fx.Option{
		opentelemetry.NewModule(),
		database.NewModule(),
		web.NewModule(),
		session.NewModule(),
		translation.NewModule(),
		validation.NewModule(),
		fx.Invoke(func(*http.Server) {}),
	}

	options = append(options, opts...)

	if err := fx.ValidateApp(options...); err != nil {
		panic(err)
	}

	fx.New(options...).Run()
}
