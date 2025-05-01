package usermgt

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"internal.snowdrop/framework/opentelemetry"
)

type CreateUserDTO struct {
	Username    string
	Email       string
	RawPassword string
}

type UserServiceImpl struct {
	structName     string
	userRepository UserRepository
	tracer         trace.Tracer
}

var _ UserService = (*UserServiceImpl)(nil)

func NewUserService(userRepository UserRepository, tracer trace.Tracer) UserServiceImpl {
	userServiceImpl := UserServiceImpl{}
	userServiceImpl.structName = reflect.TypeOf(userServiceImpl).Name()
	userServiceImpl.userRepository = userRepository
	userServiceImpl.tracer = tracer

	return userServiceImpl
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
	spanCtx, span := u.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(u.structName, "Login"),
		trace.WithAttributes(attribute.String("args[1].username", username)),
		trace.WithAttributes(attribute.String("args[2].password", "[redacted]")),
	)
	defer span.End()

	userModel, err := u.userRepository.GetUserByUsername(spanCtx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(password)); err != nil {
		return nil, err
	}

	return lo.ToPtr(ToUserDTO(*userModel)), nil
}

func (u UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*UserDTO, error) {
	spanCtx, span := u.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(u.structName, "GetUserByID"),
		trace.WithAttributes(attribute.Stringer("args[1].id", id)),
	)
	defer span.End()

	model, err := u.userRepository.GetUserByID(spanCtx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return lo.ToPtr(ToUserDTO(*model)), nil
}
