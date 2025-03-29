package usermgt

import (
	"context"

	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserDTO struct {
	Username    string
	Email       string
	RawPassword string
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

func (u UserServiceImpl) CreateUser(
	ctx context.Context,
	createUserDTO CreateUserDTO,
) (*UserDTO, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(createUserDTO.RawPassword),
		bcrypt.DefaultCost,
	)
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

	return lo.ToPtr(ToUserDTO(*userModel)), nil
}

func (u UserServiceImpl) Login(
	ctx context.Context,
	username string,
	password string,
) (*UserDTO, error) {
	userModel, err := u.userRepository.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(password)); err != nil {
		return nil, err
	}

	return lo.ToPtr(ToUserDTO(*userModel)), nil
}
