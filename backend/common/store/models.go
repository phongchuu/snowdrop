package store

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type UserModel struct {
	ID       uuid.UUID `gorm:"id"`
	Username string    `gorm:"username"`
	Email    string    `gorm:"email"`
	Password string    `gorm:"password"`

	// Readonly fields
	CreatedAt time.Time      `gorm:"created_at;->"`
	CreatedBy string         `gorm:"created_by;->"`
	UpdatedAt sql.NullTime   `gorm:"updated_at;->"`
	UpdatedBy sql.NullString `gorm:"updated_by;->"`
}

func (model UserModel) TableName() string {
	return "users"
}

func (model UserModel) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", model.ID.String()),
		slog.String("username", model.Username),
		slog.String("email", model.Email),
		slog.String("password", "[redacted]"),
		slog.Time("created_at", model.CreatedAt),
		slog.String("created_by", model.CreatedBy),
		lo.Ternary(
			model.UpdatedAt.Valid,
			slog.Time("updated_at", model.UpdatedAt.Time),
			slog.Any("updated_at", nil),
		),
		lo.Ternary(
			model.UpdatedBy.Valid,
			slog.String("updated_by", model.UpdatedBy.String),
			slog.Any("updated_by", nil),
		),
	)
}
