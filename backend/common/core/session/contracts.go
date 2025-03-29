package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"internal.snowdrop/common/store"
)

// Repository defines the database operations needed by the session manager.
type Repository interface {
	GetSession(ctx context.Context, sessionID string) (*store.SessionModel, error)
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		expiresAt time.Time,
	) (*store.SessionModel, error)
	UpdateSession(ctx context.Context, sessionID string, lastAccessedAt, expiresAt time.Time) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) error
}
