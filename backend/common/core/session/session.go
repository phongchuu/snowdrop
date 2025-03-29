package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"internal.snowdrop/common/core/database"
	"internal.snowdrop/common/store"
)

type ctxKey int8

// Config holds session configuration.
type Config struct {
	InactivityTimeout time.Duration
	MaximumTimeout    time.Duration
	CookieName        string
	SecureCookie      bool
	HTTPOnlyCookie    bool
	SameSite          http.SameSite
}

// Manager handles session operations.
type Manager struct {
	config        Config
	txMgr         database.TransactionManager
	repository    Repository
	cleanupTicker *time.Ticker
}

type ManagerParams struct {
	fx.In
	TransactionManager database.TransactionManager
	Repository         Repository
}

const sessionCtxKey ctxKey = 0

// NewManager creates a new Manager instance with default session configuration values and starts a background cleanup routine to remove expired sessions.
// It sets inactivity and maximum timeout durations, cookie parameters, and initializes the session repository and transaction manager using the provided ManagerParams.
func NewManager(p ManagerParams) *Manager {
	m := &Manager{
		config: Config{
			InactivityTimeout: 7 * 24 * time.Hour,
			MaximumTimeout:    30 * 24 * time.Hour,
			CookieName:        "session_id",
			SecureCookie:      false,
			HTTPOnlyCookie:    true,
			SameSite:          http.SameSiteLaxMode,
		},
		repository: p.Repository,
		txMgr:      p.TransactionManager,
	}

	// Start cleanup routine
	m.cleanupTicker = time.NewTicker(5 * time.Minute)
	go m.cleanupExpiredSessions()

	return m
}

func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)

			return
		}

		session, err := m.getOrCreateSession(w, r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		ctx := context.WithValue(r.Context(), sessionCtxKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Manager) GetSessionID(_ http.ResponseWriter, r *http.Request) (*string, error) {
	cookie, err := r.Cookie(m.config.CookieName)
	if err != nil {
		return nil, err
	}

	return &cookie.Value, nil
}

func (m *Manager) GetCurrentSession(r *http.Request) (*store.SessionModel, error) {
	if session, ok := r.Context().Value(sessionCtxKey).(*store.SessionModel); ok {
		return session, nil
	}

	return nil, ErrNoSessionInContext
}

func (m *Manager) UpgradeSession(
	w http.ResponseWriter,
	r *http.Request,
	userID uuid.UUID,
) (*store.SessionModel, error) {
	currentSession, getCurrentSessionErr := m.GetCurrentSession(r)
	if getCurrentSessionErr != nil {
		return nil, fmt.Errorf("cannot get the current session: %w", getCurrentSessionErr)
	}

	session, err := database.RunTxWithData(
		r.Context(),
		m.txMgr,
		func(ctx context.Context, _ *sql.Tx) (*store.SessionModel, error) {
			if err := m.repository.DeleteSession(ctx, currentSession.ID); err != nil {
				return nil, err
			}

			session, err := m.CreateSession(ctx, userID)
			if err != nil {
				return nil, err
			}

			return session, nil
		},
	)
	if err != nil {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.config.CookieName,
		Value:    session.ID,
		Path:     "/",
		Secure:   m.config.SecureCookie,
		HttpOnly: m.config.HTTPOnlyCookie,
		SameSite: m.config.SameSite,
		Expires:  session.ExpiresAt,
	})

	return session, nil
}

func (m *Manager) getOrCreateSession(
	w http.ResponseWriter,
	r *http.Request,
) (*store.SessionModel, error) {
	sessionID, getSessionIDErr := m.GetSessionID(w, r)
	if getSessionIDErr != nil && !errors.Is(getSessionIDErr, http.ErrNoCookie) {
		return nil, getSessionIDErr
	}

	if sessionID != nil {
		session, err := m.GetSession(r.Context(), *sessionID)
		if err != nil {
			return nil, err
		}

		return session, nil
	}

	// Guest session
	session, err := m.CreateSession(r.Context(), uuid.Nil)
	if err != nil {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     m.config.CookieName,
		Value:    session.ID,
		Path:     "/",
		Secure:   m.config.SecureCookie,
		HttpOnly: m.config.HTTPOnlyCookie,
		SameSite: m.config.SameSite,
		Expires:  session.ExpiresAt,
	})

	return session, nil
}

func (m *Manager) GetSession(ctx context.Context, sessionID string) (*store.SessionModel, error) {
	session, err := m.repository.GetSession(ctx, sessionID)
	if err != nil || session == nil {
		return nil, err
	}

	now := time.Now()

	newExpires := now.Add(m.config.InactivityTimeout)
	if newExpires.After(session.ExpiresAt) {
		newExpires = session.ExpiresAt
	}

	err = m.repository.UpdateSession(ctx, sessionID, now, newExpires)
	if err != nil {
		return nil, err
	}

	session.LastAccessedAt = now
	session.ExpiresAt = newExpires

	return session, nil
}

func (m *Manager) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
) (*store.SessionModel, error) {
	return m.repository.CreateSession(ctx, userID, time.Now().Add(m.config.MaximumTimeout))
}

func (m *Manager) DestroySession(ctx context.Context, sessionID string) error {
	return m.repository.DeleteSession(ctx, sessionID)
}

func (m *Manager) cleanupExpiredSessions() {
	for range m.cleanupTicker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_ = m.repository.DeleteExpiredSessions(ctx) // Best effort

		cancel()
	}
}

// generateSessionID generates a new session identifier using 32 bytes of cryptographically secure random data and encodes it in URL-safe base64 format. It returns the session ID and an error if the random data generation fails.
func generateSessionID() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
