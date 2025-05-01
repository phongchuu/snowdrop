package auth_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	ut "github.com/go-playground/universal-translator"
	"github.com/google/uuid"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/text/language"
	"internal.snowdrop/common/core"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/uuidext"
	"internal.snowdrop/framework/validation"
	"internal.snowdrop/framework/web"
	mockusermgt "internal.snowdrop/testing/mocks/internal.snowdrop/common/features/usermgt"
	mocksnowdrop "internal.snowdrop/testing/mocks/internal.snowdrop/framework"
	"internal.snowdrop/testing/testutils"
)

type ProfileRouteTestSuite struct {
	testutils.ExtendedTestSuite

	noopTracer          trace.Tracer
	i18nBundle          *i18n.Bundle
	universalTranslator *ut.UniversalTranslator

	mockConfigManager  *mocksnowdrop.MockConfigManager
	mockSessionManager *mocksnowdrop.MockSessionManager
	mockUserService    *mockusermgt.MockUserService
}

//nolint:paralleltest // This test uses a database container, so it cannot run in parallel.
func TestProfileRouteTestSuite(t *testing.T) {
	suite.Run(t, new(ProfileRouteTestSuite))
}

func (s *ProfileRouteTestSuite) SetupTest() {
	s.noopTracer = noop.NewTracerProvider().Tracer("noop.tracer")

	s.mockConfigManager = mocksnowdrop.NewMockConfigManager(s.T())
	s.mockConfigManager.EXPECT().GetSupportedLanguages().Return([]language.Tag{language.English})
	s.mockConfigManager.EXPECT().GetMaxRequestSize().Return(0)
	s.mockConfigManager.EXPECT().GetEmbedResourceFolder().Return(testutils.GetResourceFS())

	i18nBundle, err := translation.NewI18nBundle(s.mockConfigManager)
	s.Require().NoError(err)
	s.i18nBundle = i18nBundle

	validator, err := validation.NewValidator()
	s.Require().NoError(err)
	s.universalTranslator = validator.UniversalTranslator

	s.mockSessionManager = mocksnowdrop.NewMockSessionManager(s.T())
	s.mockUserService = mockusermgt.NewMockUserService(s.T())
}

func (s *ProfileRouteTestSuite) TestRoute() {
	profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
		Tracer:      s.noopTracer,
		UserService: s.mockUserService,
	})

	router := web.NewRouter(web.RouteParams{
		Config:                     s.mockConfigManager,
		Logger:                     log.NewNoopLogger(),
		I18nBundle:                 s.i18nBundle,
		HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
		SessionManager:             s.mockSessionManager,
		Tracer:                     s.noopTracer,
		UniversalTranslator:        s.universalTranslator,
		GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(s.mockConfigManager),
	})

	s.True(router.Match(chi.NewRouteContext(), http.MethodGet, "/auth/profile"))
}

func (s *ProfileRouteTestSuite) TestReturnUnauthorizedForAnonymousSession() {
	profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
		Tracer:      s.noopTracer,
		UserService: s.mockUserService,
	})

	s.mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
	s.mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
		ID:             "session-id",
		UserID:         uuid.NullUUID{},
		Data:           nil,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		ExpiresAt:      time.Now().Add(time.Minute),
	}, nil)
	s.mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
		CookieName: "session-id",
	})

	router := web.NewRouter(web.RouteParams{
		Config:                     s.mockConfigManager,
		Logger:                     log.NewNoopLogger(),
		I18nBundle:                 s.i18nBundle,
		HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
		SessionManager:             s.mockSessionManager,
		Tracer:                     s.noopTracer,
		UniversalTranslator:        s.universalTranslator,
		GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(s.mockConfigManager),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(profileRoute.Method(), profileRoute.Path(), http.NoBody)

	router.ServeHTTP(w, r)

	var result response.Response[any]

	s.Require().NoError(json.NewDecoder(w.Body).Decode(&result))
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Equal(http.StatusUnauthorized, result.Status)
	s.Equal("Authentication is required to proceed with this request.", result.Message)
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}

func (s *ProfileRouteTestSuite) TestReturnUserInfoForValidSession() {
	userDTO := usermgt.UserDTO{
		ID:        uuidext.MustUUIDV7(),
		Username:  "tester",
		Email:     "tester@internal.com",
		CreatedAt: time.Now().UTC(),
		CreatedBy: uuidext.MustUUIDV7().String(),
		UpdatedAt: lo.ToPtr(time.Now().UTC()),
		UpdatedBy: lo.ToPtr(uuidext.MustUUIDV7().String()),
	}

	s.mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
	s.mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
		ID: "session-id",
		UserID: uuid.NullUUID{
			UUID:  userDTO.ID,
			Valid: true,
		},
		Data:           nil,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		ExpiresAt:      time.Now().Add(time.Minute),
	}, nil)
	s.mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
		CookieName: "session-id",
	})

	s.mockUserService.EXPECT().GetUserByID(mock.Anything, userDTO.ID).Return(&userDTO, nil)

	profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
		Tracer:      s.noopTracer,
		UserService: s.mockUserService,
	})

	router := web.NewRouter(web.RouteParams{
		Config:                     s.mockConfigManager,
		Logger:                     log.NewNoopLogger(),
		I18nBundle:                 s.i18nBundle,
		HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
		SessionManager:             s.mockSessionManager,
		Tracer:                     s.noopTracer,
		UniversalTranslator:        s.universalTranslator,
		GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(s.mockConfigManager),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(profileRoute.Method(), profileRoute.Path(), http.NoBody)

	router.ServeHTTP(w, r)

	var result response.Response[usermgt.UserDTO]

	s.Require().NoError(json.NewDecoder(w.Body).Decode(&result))
	s.Equal(http.StatusOK, w.Code)
	s.Equal(http.StatusOK, result.Status)
	s.Equal("Your request was successfully completed.", result.Message)
	s.Equal(userDTO.ID, result.Data.ID)
	s.Equal(userDTO.Username, result.Data.Username)
	s.Equal(userDTO.Email, result.Data.Email)
	s.Equal(userDTO.CreatedAt, result.Data.CreatedAt)
	s.Equal(userDTO.CreatedBy, result.Data.CreatedBy)
	s.Equal(userDTO.UpdatedAt, result.Data.UpdatedAt)
	s.Equal(userDTO.UpdatedBy, result.Data.UpdatedBy)
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}

//nolint:dupl // Fix later
func (s *ProfileRouteTestSuite) TestUnauthorizedWhenSessionValidButUserDeleted() {
	userDTO := usermgt.UserDTO{
		ID: uuidext.MustUUIDV7(),
	}

	s.mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
	s.mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
		ID: "session-id",
		UserID: uuid.NullUUID{
			UUID:  userDTO.ID,
			Valid: true,
		},
		Data:           nil,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		ExpiresAt:      time.Now().Add(time.Minute),
	}, nil)
	s.mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
		CookieName: "session-id",
	})

	s.mockUserService.EXPECT().GetUserByID(mock.Anything, userDTO.ID).Return(nil, sql.ErrNoRows)

	profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
		Tracer:      s.noopTracer,
		UserService: s.mockUserService,
	})

	router := web.NewRouter(web.RouteParams{
		Config:                     s.mockConfigManager,
		Logger:                     log.NewNoopLogger(),
		I18nBundle:                 s.i18nBundle,
		HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
		SessionManager:             s.mockSessionManager,
		Tracer:                     s.noopTracer,
		UniversalTranslator:        s.universalTranslator,
		GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(s.mockConfigManager),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(profileRoute.Method(), profileRoute.Path(), http.NoBody)

	router.ServeHTTP(w, r)

	var result response.Response[any]

	s.Require().NoError(json.NewDecoder(w.Body).Decode(&result))
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Equal(http.StatusUnauthorized, result.Status)
	s.Equal("Authentication is required to proceed with this request.", result.Message)
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}

//nolint:dupl // Fix later
func (s *ProfileRouteTestSuite) TestHandlerReturns500WhenUserInfoRetrievalFailsWithValidSession() {
	userDTO := usermgt.UserDTO{
		ID: uuidext.MustUUIDV7(),
	}

	s.mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
	s.mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
		ID: "session-id",
		UserID: uuid.NullUUID{
			UUID:  userDTO.ID,
			Valid: true,
		},
		Data:           nil,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		ExpiresAt:      time.Now().Add(time.Minute),
	}, nil)
	s.mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
		CookieName: "session-id",
	})

	s.mockUserService.EXPECT().GetUserByID(mock.Anything, userDTO.ID).Return(nil, sql.ErrConnDone)

	profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
		Tracer:      s.noopTracer,
		UserService: s.mockUserService,
	})

	router := web.NewRouter(web.RouteParams{
		Config:                     s.mockConfigManager,
		Logger:                     log.NewNoopLogger(),
		I18nBundle:                 s.i18nBundle,
		HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
		SessionManager:             s.mockSessionManager,
		Tracer:                     s.noopTracer,
		UniversalTranslator:        s.universalTranslator,
		GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(s.mockConfigManager),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(profileRoute.Method(), profileRoute.Path(), http.NoBody)

	router.ServeHTTP(w, r)

	var result response.Response[any]

	s.Require().NoError(json.NewDecoder(w.Body).Decode(&result))
	s.Equal(http.StatusInternalServerError, w.Code)
	s.Equal(http.StatusInternalServerError, result.Status)
	s.Equal("An internal server error occurred; please try again later.", result.Message)
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}
