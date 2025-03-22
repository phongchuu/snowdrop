package usermgt

import (
	"context"
)

type UserRepository interface {
	// CreateUser inserts a new user record into the database.
	CreateUser(ctx context.Context, setter UserSetter) (*UserModel, error)
}

type UserService interface {
	// CreateUser creates a new user with the provided details.
	// It hashes the user's password before saving it to the database.
	//
	// Parameters:
	//   - ctx: The context for the request, used for cancellation and deadlines.
	//   - createUserDTO: A data transfer object containing the user's details.
	//
	// Returns:
	//   - A pointer to a UserDTO containing the created user's details.
	//   - An error if the user could not be created.
	CreateUser(ctx context.Context, createUserDTO CreateUserDTO) (*UserDTO, error)
}
