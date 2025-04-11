package usermgt

import (
	"context"

	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserDTO struct {
	Username    string
	Email       string
	RawPassword string
}

type UserServiceImpl struct {
	userRepository UserRepository
	tracer         trace.Tracer
}

var _ UserService = (*UserServiceImpl)(nil)

func NewUserService(userRepository UserRepository, tracer trace.Tracer) UserServiceImpl {
	return UserServiceImpl{
		userRepository: userRepository,
		tracer:         tracer,
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
	ctx, span := u.tracer.Start(
		ctx,
		"[user.service.go] Login",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("username", username)),
		trace.WithAttributes(attribute.String("password", "[redacted]")),
	)
	defer span.End()

	userModel, err := u.userRepository.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(password)); err != nil {
		return nil, err
	}

	return lo.ToPtr(ToUserDTO(*userModel)), nil
}
