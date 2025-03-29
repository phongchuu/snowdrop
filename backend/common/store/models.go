package store

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type SessionModel struct {
	ID             string        `db:"id"`
	UserID         uuid.NullUUID `db:"user_id"`
	CreatedAt      time.Time     `db:"created_at"`
	LastAccessedAt time.Time     `db:"last_accessed_at"`
	ExpiresAt      time.Time     `db:"expires_at"`
	Data           []byte        `db:"data"`
}

type UserModel struct {
	ID        uuid.UUID      `db:"id"`
	Username  string         `db:"username"`
	Email     string         `db:"email"`
	Password  string         `db:"password"`
	CreatedAt time.Time      `db:"created_at"`
	CreatedBy string         `db:"created_by"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
	UpdatedBy sql.NullString `db:"updated_by"`
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
