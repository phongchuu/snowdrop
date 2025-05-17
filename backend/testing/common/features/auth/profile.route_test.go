package auth_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
	"internal.snowdrop/common/core"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/uuidext"
	"internal.snowdrop/framework/web"
	mockusermgt "internal.snowdrop/testing/mocks/internal.snowdrop/common/features/usermgt"
	mocksnowdrop "internal.snowdrop/testing/mocks/internal.snowdrop/framework"
	"internal.snowdrop/testing/testutils"
)

var _ = Describe("[ProfileRoute]", Label("integration"), Serial, func() {
	var (
		di                 testutils.TestContainer
		mockUserService    *mockusermgt.MockUserService
		mockSessionManager *mocksnowdrop.MockSessionManager
	)

	BeforeEach(func() {
		di = testutils.IntegrationTestSetup(GinkgoTB())

		di.MockConfigManager.EXPECT().GetSupportedLanguages().Return([]language.Tag{language.English})
		di.MockConfigManager.EXPECT().GetMaxRequestSize().Return(0)
		di.MockConfigManager.EXPECT().GetEmbedResourceFolder().Return(testutils.GetResourceFS())
		mockSessionManager = mocksnowdrop.NewMockSessionManager(GinkgoT())
		mockUserService = mockusermgt.NewMockUserService(GinkgoT())
	})

	Context("When accessing profile endpoint", func() {
		It("should return unauthorized for anonymous session", func() {
			profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
				Tracer:      di.NoopTracer,
				UserService: mockUserService,
			})

			mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
			mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
				ID:             "session-id",
				UserID:         uuid.NullUUID{},
				Data:           nil,
				CreatedAt:      time.Now(),
				LastAccessedAt: time.Now(),
				ExpiresAt:      time.Now().Add(time.Minute),
			}, nil)
			mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
				CookieName: "session-id",
			})

			router := web.NewRouter(web.RouteParams{
				Config:                     di.MockConfigManager,
				Logger:                     log.NewNoopLogger(),
				I18nBundle:                 di.I18nBundle,
				HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
				SessionManager:             mockSessionManager,
				Tracer:                     di.NoopTracer,
				UniversalTranslator:        di.UniversalTranslator,
				GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(di.MockConfigManager),
			})

			server := httptest.NewServer(router)
			defer server.Close()

			httpexpecter := httpexpect.Default(GinkgoT(), server.URL)
			response := httpexpecter.GET("/auth/profile").Expect()
			response.Status(http.StatusUnauthorized)
			Expect(response.JSON().Object().Raw()).To(MatchAllKeys(Keys{
				"status":    BeNumerically("==", http.StatusUnauthorized),
				"message":   Equal("Authentication is required to proceed with this request."),
				"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
			}))
		})

		It("should return user information for valid session", func() {
			userDTO := usermgt.UserDTO{
				ID:        uuidext.MustUUIDV7(),
				Username:  "tester",
				Email:     "tester@internal.com",
				CreatedAt: time.Now().UTC(),
				CreatedBy: uuidext.MustUUIDV7().String(),
				UpdatedAt: lo.ToPtr(time.Now().UTC()),
				UpdatedBy: lo.ToPtr(uuidext.MustUUIDV7().String()),
			}

			mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
			mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
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
			mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
				CookieName: "session-id",
			})

			mockUserService.EXPECT().GetUserByID(mock.Anything, userDTO.ID).Return(&userDTO, nil)

			profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
				Tracer:      di.NoopTracer,
				UserService: mockUserService,
			})

			router := web.NewRouter(web.RouteParams{
				Config:                     di.MockConfigManager,
				Logger:                     log.NewNoopLogger(),
				I18nBundle:                 di.I18nBundle,
				HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
				SessionManager:             mockSessionManager,
				Tracer:                     di.NoopTracer,
				UniversalTranslator:        di.UniversalTranslator,
				GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(di.MockConfigManager),
			})

			server := httptest.NewServer(router)
			defer server.Close()

			httpexpecter := httpexpect.Default(GinkgoT(), server.URL)

			response := httpexpecter.GET("/auth/profile").Expect()
			response.Status(http.StatusOK)
			Expect(response.JSON().Object().Raw()).To(MatchAllKeys(Keys{
				"status":    BeNumerically("==", http.StatusOK),
				"message":   Equal("Your request was successfully completed."),
				"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				"data": MatchKeys(IgnoreExtras, Keys{
					"id":        Equal(userDTO.ID.String()),
					"username":  Equal(userDTO.Username),
					"email":     Equal(userDTO.Email),
					"createdAt": Equal(userDTO.CreatedAt.Format(time.RFC3339Nano)),
					"createdBy": Equal(userDTO.CreatedBy),
					"updatedAt": Equal(userDTO.UpdatedAt.Format(time.RFC3339Nano)),
					"updatedBy": Equal(*userDTO.UpdatedBy),
				}),
			}))
		})

		DescribeTable(
			"should handle user service errors appropriately",
			func(serviceError error, expectedCode int, expectedMsg string) {
				userID := uuidext.MustUUIDV7()

				mockSessionManager.EXPECT().GetSessionIDFromCookie(mock.Anything).Return(lo.ToPtr("random-id"), nil)
				mockSessionManager.EXPECT().GetSession(mock.Anything, "random-id").Return(&snowdrop.SessionModel{
					ID: "session-id",
					UserID: uuid.NullUUID{
						UUID:  userID,
						Valid: true,
					},
					Data:           nil,
					CreatedAt:      time.Now(),
					LastAccessedAt: time.Now(),
					ExpiresAt:      time.Now().Add(time.Minute),
				}, nil)
				mockSessionManager.EXPECT().GetConfig().Return(snowdrop.SessionConfig{
					CookieName: "session-id",
				})

				mockUserService.EXPECT().GetUserByID(mock.Anything, userID).Return(nil, serviceError)

				profileRoute := auth.NewProfileRoute(auth.ProfileRouteParams{
					Tracer:      di.NoopTracer,
					UserService: mockUserService,
				})

				router := web.NewRouter(web.RouteParams{
					Config:                     di.MockConfigManager,
					Logger:                     log.NewNoopLogger(),
					I18nBundle:                 di.I18nBundle,
					HTTPRoutes:                 []snowdrop.HTTPHandler{profileRoute},
					SessionManager:             mockSessionManager,
					Tracer:                     di.NoopTracer,
					UniversalTranslator:        di.UniversalTranslator,
					GetPreferredUserLanguageFn: core.NewGetPreferredUserLanguageFn(di.MockConfigManager),
				})

				server := httptest.NewServer(router)
				defer server.Close()

				httpexpecter := httpexpect.Default(GinkgoT(), server.URL)

				response := httpexpecter.GET("/auth/profile").Expect()
				response.Status(expectedCode)
				Expect(response.JSON().Object().Raw()).To(MatchAllKeys(Keys{
					"status":    BeNumerically("==", expectedCode),
					"message":   Equal(expectedMsg),
					"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				}))
			},
			Entry(
				"when user is deleted",
				sql.ErrNoRows,
				http.StatusUnauthorized,
				"Authentication is required to proceed with this request.",
			),
			Entry(
				"when database error occurs",
				sql.ErrConnDone,
				http.StatusInternalServerError,
				"An internal server error occurred; please try again later.",
			),
		)
	})
})
