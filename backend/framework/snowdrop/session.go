package snowdrop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SessionModel struct {
	ID             string           `db:"id"`
	UserID         uuid.NullUUID    `db:"user_id"`
	Data           *json.RawMessage `db:"data"`
	CreatedAt      time.Time        `db:"created_at"`
	LastAccessedAt time.Time        `db:"last_accessed_at"`
	ExpiresAt      time.Time        `db:"expires_at"`
}

// Config holds session configuration.
type SessionConfig struct {
	InactivityTimeout time.Duration
	MaximumTimeout    time.Duration
	CookieName        string
	SecureCookie      bool
	HTTPOnlyCookie    bool
	SameSite          http.SameSite
}

type SessionManager interface {
	GetConfig() SessionConfig
	GetSessionIDFromCookie(r *http.Request) (*string, error)
	UpgradeSession(
		w http.ResponseWriter,
		r *http.Request,
		userID uuid.UUID,
	) (*SessionModel, error)
	GetSession(
		ctx context.Context,
		sessionID string,
	) (*SessionModel, error)
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		data map[string]any,
	) (*SessionModel, error)
	DestroySession(ctx context.Context, sessionID string) error
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
		data map[string]any,
		expiresAt time.Time,
	) (*SessionModel, error)

	// UpdateSession updates the last accessed time and expiration time for a session.
	UpdateSession(ctx context.Context, sessionID string, lastAccessedAt, expiresAt time.Time) error

	// DeleteSession removes a specific session from storage.
	DeleteSession(ctx context.Context, sessionID string) error

	// DeleteExpiredSessions removes all sessions that have passed their expiration time.
	DeleteExpiredSessions(ctx context.Context) error
}

type contextKey int

const ctxID contextKey = iota

var ErrNoSession = errors.New("there is no snowdrop.SessionModel in the given context")

func WithSession(r *http.Request, session SessionModel) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxID, session))
}

func GetCurrentSession(r *http.Request) (*SessionModel, error) {
	tracer, err := GetTracer(r)
	if err != nil {
		return nil, err
	}

	spanCtx, span := tracer.Start(r.Context(), "snowdrop.GetCurrentSession")
	defer span.End()

	return GetCurrentSessionFromCtx(spanCtx)
}

func GetCurrentSessionFromCtx(ctx context.Context) (*SessionModel, error) {
	tracer, err := GetTracerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	spanCtx, span := tracer.Start(ctx, "snowdrop.GetCurrentSessionFromCtx")
	defer span.End()

	if session, ok := spanCtx.Value(ctxID).(SessionModel); ok {
		span.SetStatus(codes.Ok, "")

		return &session, nil
	}

	span.RecordError(ErrNoSession, trace.WithStackTrace(true))

	return nil, ErrNoSession
}
