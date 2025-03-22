package usermgt

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgtype"
	"internal.snowdrop/common/core/database"
)

type UserModel struct {
	ID        pgtype.UUID        `db:"id"`
	Username  string             `db:"username"`
	Email     string             `db:"email"`
	Password  string             `db:"password"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
	CreatedBy string             `db:"created_by"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`
	UpdatedBy sql.NullString     `db:"updated_by"`
}

type UserSetter struct {
	Username string
	Email    string
	Password string
}

func NewUserModel() UserModel {
	return UserModel{
		ID: database.GenerateUUIDv7PgType(),
	}
}
