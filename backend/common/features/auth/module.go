package auth

import (
	"go.uber.org/fx"
	"internal.snowdrop/common/core/web"
)

// NewModule creates an fx.Option that configures the AuthModule by registering HTTP routes for authentication.
// It sets up the module with two providers: one for the registration route (via NewRegisterRoute)
// and one for the login route (via NewLoginRoute).
func NewModule() fx.Option {
	return fx.Module(
		"AuthModule",
		fx.Provide(web.HTTPRoute(NewRegisterRoute)),
		fx.Provide(web.HTTPRoute(NewLoginRoute)),
	)
}
