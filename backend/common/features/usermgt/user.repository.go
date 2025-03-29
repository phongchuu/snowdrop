package usermgt

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"go.uber.org/fx"
	"internal.snowdrop/common/core/database"
	"internal.snowdrop/common/core/uuidext"
	"internal.snowdrop/common/store"
)

type UserRepositoryImpl struct {
	txMgr database.TransactionManager
}

type UserRepositoryParams struct {
	fx.In
	TransactionManager database.TransactionManager
}

var _ UserRepository = (*UserRepositoryImpl)(nil)

// NewUserRepository returns a new UserRepositoryImpl configured with the TransactionManager provided in UserRepositoryParams.
func NewUserRepository(p UserRepositoryParams) UserRepositoryImpl {
	return UserRepositoryImpl{
		txMgr: p.TransactionManager,
	}
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
