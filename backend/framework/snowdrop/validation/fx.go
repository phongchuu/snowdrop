package validation

import "go.uber.org/fx"

func NewModule() fx.Option {
	return fx.Module(
		"ValidationModule",
		fx.Provide(NewValidator),
	)
}
