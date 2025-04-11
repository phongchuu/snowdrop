package snowdrop

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionModel struct {
	ID             string        `db:"id"`
	UserID         uuid.NullUUID `db:"user_id"`
	CreatedAt      time.Time     `db:"created_at"`
	LastAccessedAt time.Time     `db:"last_accessed_at"`
	ExpiresAt      time.Time     `db:"expires_at"`
	Data           []byte        `db:"data"`
}

// SessionRepository manages user sessions with CRUD operations and expiration cleanup.
type SessionRepository interface {
	// GetSession retrieves a session by its ID. Returns the session model if found,
	// or an error if the session doesn't exist or if there's a database error.
	GetSession(ctx context.Context, sessionID string) (*SessionModel, error)

	// CreateSession creates a new session for a user with specified expiration time.
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		expiresAt time.Time,
	) (*SessionModel, error)

	// UpdateSession updates the last accessed time and expiration time for a session.
	UpdateSession(ctx context.Context, sessionID string, lastAccessedAt, expiresAt time.Time) error

	// DeleteSession removes a specific session from storage.
	DeleteSession(ctx context.Context, sessionID string) error

	// DeleteExpiredSessions removes all sessions that have passed their expiration time.
	DeleteExpiredSessions(ctx context.Context) error
}
