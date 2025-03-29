package session

import "go.uber.org/fx"

func NewModule() fx.Option {
	return fx.Module(
		"SessionModule",
		fx.Provide(
			NewManager,
			fx.Annotate(
				NewPostgresRepository,
				fx.As(new(Repository)),
			),
		),
	)
}
