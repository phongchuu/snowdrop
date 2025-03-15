package trans

import "go.uber.org/fx"

func NewModule() fx.Option {
	return fx.Module(
		"I18nModule",
		fx.Provide(NewI18nBundle),
	)
}
