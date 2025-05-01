package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/opentelemetry"
)

// Manager handles session operations.
type Manager struct {
	structName    string
	config        snowdrop.SessionConfig
	txMgr         snowdrop.TransactionManager
	repository    snowdrop.SessionRepository
	cleanupTicker *time.Ticker
	tracer        trace.Tracer
}

type ManagerParams struct {
	fx.In
	TransactionManager snowdrop.TransactionManager
	Repository         snowdrop.SessionRepository
	Tracer             trace.Tracer
}

var _ snowdrop.SessionManager = (*Manager)(nil)

func NewManager(p ManagerParams) Manager {
	m := Manager{}
	m.structName = reflect.TypeOf(m).Name()
	m.config = snowdrop.SessionConfig{
		InactivityTimeout: 7 * 24 * time.Hour,
		MaximumTimeout:    30 * 24 * time.Hour,
		CookieName:        "session_id",
		SecureCookie:      false,
		HTTPOnlyCookie:    true,
		SameSite:          http.SameSiteLaxMode,
	}
	m.repository = p.Repository
	m.txMgr = p.TransactionManager
	m.tracer = p.Tracer

	// Start cleanup routine
	m.cleanupTicker = time.NewTicker(5 * time.Minute)
	go m.cleanupExpiredSessions()

	return m
}

func (m Manager) GetConfig() snowdrop.SessionConfig {
	return m.config
}

func (m Manager) GetSessionIDFromCookie(r *http.Request) (*string, error) {
	_, span := m.tracer.Start(
		r.Context(),
		opentelemetry.BuildSpanName(m.structName, "GetSessionIDFromCookie"),
	)
	defer span.End()

	cookie, err := r.Cookie(m.config.CookieName)
	if err != nil {
		return nil, err
	}

	return &cookie.Value, nil
}

func (m Manager) UpgradeSession(
	w http.ResponseWriter,
	r *http.Request,
	userID uuid.UUID,
) (*snowdrop.SessionModel, error) {
	currentSession, getCurrentSessionErr := snowdrop.GetCurrentSession(r)
	if getCurrentSessionErr != nil {
		return nil, fmt.Errorf("cannot get the current session: %w", getCurrentSessionErr)
	}

	session, err := database.RunTxWithData(
		r.Context(),
		m.txMgr,
		func(ctx context.Context, _ *sql.Tx) (*snowdrop.SessionModel, error) {
			if err := m.repository.DeleteSession(ctx, currentSession.ID); err != nil {
				return nil, err
			}

			session, err := m.CreateSession(ctx, userID, nil)
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

func (m Manager) GetSession(
	ctx context.Context,
	sessionID string,
) (*snowdrop.SessionModel, error) {
	spanCtx, span := m.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(m.structName, "GetSession"),
	)
	defer span.End()

	session, err := m.repository.GetSession(spanCtx, sessionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	newExpires := now.Add(m.config.InactivityTimeout)
	if newExpires.After(session.ExpiresAt) {
		newExpires = session.ExpiresAt
	}

	err = m.repository.UpdateSession(spanCtx, sessionID, now, newExpires)
	if err != nil {
		return nil, err
	}

	session.LastAccessedAt = now
	session.ExpiresAt = newExpires

	span.SetStatus(codes.Ok, "")

	return session, nil
}

func (m Manager) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	data map[string]any,
) (*snowdrop.SessionModel, error) {
	spanCtx, span := m.tracer.Start(
		ctx,
		opentelemetry.BuildSpanName(m.structName, "CreateSession"),
	)
	defer span.End()

	return m.repository.CreateSession(spanCtx, userID, data, time.Now().Add(m.config.MaximumTimeout))
}

func (m Manager) DestroySession(ctx context.Context, sessionID string) error {
	return m.repository.DeleteSession(ctx, sessionID)
}

func (m Manager) cleanupExpiredSessions() {
	for range m.cleanupTicker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_ = m.repository.DeleteExpiredSessions(ctx) // Best effort

		cancel()
	}
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
