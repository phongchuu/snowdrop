package command

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/opentelemetry"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/validation"
	"internal.snowdrop/framework/web"
)

func defaultModules() []fx.Option {
	return []fx.Option{
		opentelemetry.NewModule(),
		database.NewModule(),
		web.NewModule(),
		session.NewModule(),
		translation.NewModule(),
		validation.NewModule(),
	}
}

// Start initializes and starts the Snowdrop application with the provided options.
func Start(opts []fx.Option) {
	time.Local = time.UTC

	modules := defaultModules()
	modules = append(modules, fx.Invoke(func(*http.Server) {}))
	modules = append(modules, opts...)

	if err := fx.ValidateApp(modules...); err != nil {
		panic(err)
	}

	fx.New(modules...).Run()
}

func StartTest(tb testing.TB, opts []fx.Option) *fxtest.App {
	time.Local = time.UTC

	modules := defaultModules()
	modules = append(modules, opts...)

	return fxtest.New(tb, modules...)
}
