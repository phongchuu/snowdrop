package usermgt

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"internal.snowdrop/common/store"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/opentelemetry"
	"internal.snowdrop/framework/uuidext"
)

type UserRepositoryImpl struct {
	structName string
	txMgr      snowdrop.TransactionManager
	tracer     trace.Tracer
}

type UserRepositoryParams struct {
	fx.In
	TransactionManager snowdrop.TransactionManager
	Tracer             trace.Tracer
}

var _ UserRepository = (*UserRepositoryImpl)(nil)

func NewUserRepository(p UserRepositoryParams) UserRepositoryImpl {
	userRepositoryImpl := UserRepositoryImpl{}
	userRepositoryImpl.structName = reflect.TypeOf(userRepositoryImpl).Name()
	userRepositoryImpl.txMgr = p.TransactionManager
	userRepositoryImpl.tracer = p.Tracer

	return userRepositoryImpl
}

// CreateUser implements UserRepository.
func (u UserRepositoryImpl) CreateUser(
	ctx context.Context,
	params UserSetter,
) (*store.UserModel, error) {
	return database.RunTxWithData(
		ctx,
		u.txMgr,
		func(ctx context.Context, tx *sql.Tx) (*store.UserModel, error) {
			model := store.UserModel{
				ID:       uuidext.MustUUIDV7(),
				Username: params.Username,
				Email:    params.Email,
				Password: params.Password,
			}

			if err := sqlscan.Get(
				ctx,
				tx,
				&model,
				`INSERT INTO public.users(id, username, email, password)
            VALUES ($1, $2, $3, $4)
            RETURNING *`,
				model.ID,
				model.Username,
				model.Email,
				model.Password,
			); err != nil {
				return nil, err
			}

			return &model, nil
		},
	)
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
		func(ctx context.Context, tx *sql.Tx) (*store.UserModel, error) {
			var userModel store.UserModel

			err := sqlscan.Get(
				ctx,
				tx,
				&userModel,
				`SELECT id, username, email, created_at, created_by, updated_at, updated_by
                FROM public.users WHERE id = $1`,
				id,
			)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return nil, err
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
		func(ctx context.Context, tx *sql.Tx) (*store.UserModel, error) {
			var userModel store.UserModel

			if err := sqlscan.Get(
				ctx,
				tx,
				&userModel,
				`SELECT id, username, email, password, created_at, created_by, updated_at, updated_by
            FROM public.users
            WHERE username = $1`,
				username,
			); err != nil {
				return nil, err
			}

			return &userModel, nil
		},
	)
}
