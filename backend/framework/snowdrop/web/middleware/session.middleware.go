package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"

	"github.com/google/uuid"
	"github.com/samber/lo"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/response"
)

func NewSessionMiddleware(
	sessionManager snowdrop.SessionManager,
	skippedRoutes []string,
) func(next http.Handler) http.Handler {
	skippedRouteRegexPatterns := lo.Map(skippedRoutes, func(skippedRoute string, _ int) *regexp.Regexp {
		return regexp.MustCompile(skippedRoute)
	})

	isSkippedRoute := func(r *http.Request) bool {
		return slices.ContainsFunc(skippedRouteRegexPatterns, func(pattern *regexp.Regexp) bool {
			route := fmt.Sprintf("%v %v", r.Method, r.URL.Path)

			return pattern.MatchString(route)
		})
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := snowdrop.MustGetTracer(r)

			spanCtx, span := tracer.Start(r.Context(), "SessionMiddleware")
			defer span.End()

			req := r.WithContext(spanCtx)

			if isSkippedRoute(req) {
				span.AddEvent("This route is skipped.")
				next.ServeHTTP(w, req)

				return
			}

			sessionModel, err := getOrCreateSession(req, sessionManager)
			if err != nil {
				response.NewBuilder(w, req).
					Status(http.StatusInternalServerError).
					JSON()

				return
			}

			sessionConfig := sessionManager.GetConfig()
			http.SetCookie(w, &http.Cookie{
				Name:     sessionConfig.CookieName,
				Value:    sessionModel.ID,
				Path:     "/",
				Secure:   sessionConfig.SecureCookie,
				HttpOnly: sessionConfig.HTTPOnlyCookie,
				SameSite: sessionConfig.SameSite,
				Expires:  sessionModel.ExpiresAt,
			})
			next.ServeHTTP(w, snowdrop.WithSession(req, *sessionModel))
		})
	}
}

func getOrCreateSession(r *http.Request, sessionManager snowdrop.SessionManager) (*snowdrop.SessionModel, error) {
	sessionID, err := sessionManager.GetSessionIDFromCookie(r)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, err
	}

	if sessionID == nil || len(*sessionID) == 0 {
		return sessionManager.CreateSession(r.Context(), uuid.Nil, nil)
	}

	model, err := sessionManager.GetSession(r.Context(), *sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sessionManager.CreateSession(r.Context(), uuid.Nil, nil)
		}

		return nil, err
	}

	return model, nil
}
