package usermgt

import (
	"context"
	"database/sql"

	"github.com/samber/lo"
	"internal.snowdrop/common/core/database"
)

type UserRepositoryImpl struct {
	db *sql.DB
	tx *sql.Tx
}

var _ UserRepository = (*UserRepositoryImpl)(nil)

func NewUserRepository(db *sql.DB) UserRepositoryImpl {
	return UserRepositoryImpl{
		db: db,
		tx: nil,
	}
}

func (u UserRepositoryImpl) WithTx(tx *sql.Tx) UserRepositoryImpl {
	return UserRepositoryImpl{
		db: nil,
		tx: tx,
	}
}

// CreateUser implements UserRepository.
func (u UserRepositoryImpl) CreateUser(
	ctx context.Context,
	setter UserSetter,
) (*UserModel, error) {
	tx, owner, err := database.GetTransaction(ctx, lo.Ternary[database.Database](u.tx == nil, u.db, u.tx))
	if err != nil {
		return nil, err
	}

	if owner {
		defer func() {
			_ = database.AutoRollbackOrCommit(tx, err)
		}()
	}

	model := NewUserModel()
	model.Username = setter.Username
	model.Email = setter.Email
	model.Password = setter.Password

	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO public.users(id, username, email, password) VALUES ($1, $2, $3, $4) RETURNING created_at, created_by`,
		model.ID,
		model.Username,
		model.Email,
		model.Password,
	).Scan(&model.CreatedAt, &model.CreatedBy); err != nil {
		return nil, err
	}

	return &model, nil
}
