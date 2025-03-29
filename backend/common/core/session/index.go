package session

import "go.uber.org/fx"

// NewModule returns an fx.Option that constructs the "SessionModule" for session management.
// It registers a session manager from NewManager and a PostgreSQL repository from NewPostgresRepository,
// with the repository annotated to implement the Repository interface.
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
