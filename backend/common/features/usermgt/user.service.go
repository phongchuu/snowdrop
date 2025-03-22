package usermgt

import (
	"context"
	"time"

	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserDTO struct {
	Username    string
	Email       string
	RawPassword string
}

type UserDTO struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *time.Time `json:"updatedAt"`
	UpdatedBy *string    `json:"updatedBy"`
}

type UserServiceImpl struct {
	userRepository UserRepository
}

var _ UserService = (*UserServiceImpl)(nil)

func NewUserService(userRepository UserRepository) UserServiceImpl {
	return UserServiceImpl{
		userRepository: userRepository,
	}
}

func (u UserServiceImpl) CreateUser(ctx context.Context, createUserDTO CreateUserDTO) (*UserDTO, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createUserDTO.RawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userModel, err := u.userRepository.CreateUser(ctx, UserSetter{
		Username: createUserDTO.Username,
		Email:    createUserDTO.Email,
		Password: string(hashedPassword),
	})

	if err != nil {
		return nil, err
	}

	return &UserDTO{
		ID:        userModel.ID.String(),
		Username:  userModel.Username,
		Email:     userModel.Email,
		CreatedAt: userModel.CreatedAt.Time,
		CreatedBy: userModel.CreatedBy,
		UpdatedAt: lo.Ternary(userModel.UpdatedAt.Valid, &userModel.UpdatedAt.Time, nil),
		UpdatedBy: lo.Ternary(userModel.UpdatedBy.Valid, &userModel.UpdatedBy.String, nil),
	}, nil
}
