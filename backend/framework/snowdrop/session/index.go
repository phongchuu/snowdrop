package session

import (
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

func NewModule() fx.Option {
	return fx.Module(
		"SessionModule",
		fx.Provide(
			fx.Annotate(
				NewManager,
				fx.As(new(snowdrop.SessionManager)),
			),
			fx.Annotate(
				NewPostgresRepository,
				fx.As(new(snowdrop.SessionRepository)),
			),
		),
	)
}
