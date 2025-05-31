package usermgt

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"internal.snowdrop/common/store"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/opentelemetry"
	"internal.snowdrop/framework/uuidext"
)

type UserRepositoryImpl struct {
	structName string
	txMgr      snowdrop.TransactionManager
	gormDB     *gorm.DB
	tracer     trace.Tracer
}

type UserRepositoryParams struct {
	fx.In
	TransactionManager snowdrop.TransactionManager
	GormDB             *gorm.DB
	Tracer             trace.Tracer
}

var _ UserRepository = (*UserRepositoryImpl)(nil)

func NewUserRepository(p UserRepositoryParams) UserRepositoryImpl {
	userRepositoryImpl := UserRepositoryImpl{}
	userRepositoryImpl.structName = reflect.TypeOf(userRepositoryImpl).Name()
	userRepositoryImpl.txMgr = p.TransactionManager
	userRepositoryImpl.tracer = p.Tracer
	userRepositoryImpl.gormDB = p.GormDB

	return userRepositoryImpl
}

// CreateUser implements UserRepository.
func (u UserRepositoryImpl) CreateUser(
	ctx context.Context,
	params UserSetter,
) (*store.UserModel, error) {
	model := store.UserModel{
		ID:       uuidext.MustUUIDV7(),
		Username: params.Username,
		Email:    params.Email,
		Password: params.Password,
	}

	result := u.gormDB.WithContext(ctx).Clauses(clause.Returning{}).Create(&model)
	if result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func (u UserRepositoryImpl) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (*store.UserModel, error) {
	spanCtx, span := u.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(u.structName, "GetUserByID"),
		trace.WithAttributes(attribute.Stringer("args[1].id", id)),
	)
	defer span.End(trace.WithStackTrace(true))

	return database.RunTxWithData(
		spanCtx,
		u.txMgr,
		func(ctx context.Context, tx *gorm.DB) (*store.UserModel, error) {
			var userModel store.UserModel

			result := tx.First(&userModel, "id = ?", id)
			if result.Error != nil {
				span.RecordError(result.Error)
				span.SetStatus(codes.Error, result.Error.Error())

				return nil, result.Error
			}

			span.SetStatus(codes.Ok, "")

			return &userModel, nil
		},
	)
}

func (u UserRepositoryImpl) GetUserByUsername(
	ctx context.Context,
	username string,
) (*store.UserModel, error) {
	return database.RunTxWithData(
		ctx,
		u.txMgr,
		func(ctx context.Context, tx *gorm.DB) (*store.UserModel, error) {
			var userModel store.UserModel

			result := tx.First(&userModel, "username = ?", username)

			if result.Error != nil {
				return nil, result.Error
			}

			return &userModel, nil
		},
	)
}
