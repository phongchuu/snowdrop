package session

import (
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

func NewModule() fx.Option {
	return fx.Module(
		"SessionModule",
		fx.Provide(
			NewManager,
			fx.Annotate(
				NewPostgresRepository,
				fx.As(new(snowdrop.SessionRepository)),
			),
		),
	)
}
