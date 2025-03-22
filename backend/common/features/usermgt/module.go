package usermgt

import "go.uber.org/fx"

func NewModule() fx.Option {
	return fx.Module(
		"UsermgtModule",
		fx.Provide(
			fx.Annotate(NewUserRepository, fx.As(new(UserRepository))),
			fx.Annotate(NewUserService, fx.As(new(UserService))),
		),
	)
}
