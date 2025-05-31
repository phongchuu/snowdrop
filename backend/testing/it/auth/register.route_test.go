package auth_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gavv/httpexpect/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/spf13/cast"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"internal.snowdrop/common/core"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	snowdropCommand "internal.snowdrop/framework/command"
	"internal.snowdrop/testing/testutils"
	"internal.snowdrop/testing/testutils/extmatchers"
)

var _ = Describe("[RegisterRoute]", Label("integration"), Serial, func() {
	var (
		app          *fxtest.App
		server       *httptest.Server
		httpexpecter *httpexpect.Expect
	)

	BeforeEach(func(ctx SpecContext) {
		pgContainer := testutils.StartPostgresContainer(ctx)

		GinkgoT().Setenv("APP_DB_DSN", pgContainer.MustConnectionString(ctx))
		GinkgoT().Setenv("APP_DEFAULT_SYSTEM_PASSWORD", "Keep!t5ecret")

		app = snowdropCommand.StartTest(
			GinkgoTB(),
			[]fx.Option{
				core.NewModule(core.ModuleConfig{
					Version:              "0.0.1.development",
					Revision:             "0.0.1.development",
					EmbedResourcesFolder: testutils.GetResourceFS(),
				}),
				usermgt.NewModule(),
				auth.NewModule(),
				fx.Invoke(func(router http.Handler) {
					server = httptest.NewServer(router)
					httpexpecter = httpexpect.Default(GinkgoTB(), server.URL)
				}),
			},
		)

		app.RequireStart()
	})

	AfterEach(func() {
		server.Close()
		app.RequireStop()
	})

	Context("with valid registration data", func() {
		It("should successfully create a new user", func() {
			response := httpexpecter.POST("/auth/register").WithJSON(map[string]string{
				"username": "tester",
				"email":    "tester@internal.com",
				"password": "Keep!T5ecret",
			}).Expect()

			response.Status(http.StatusCreated)
			Expect(response.JSON().Object().Raw()).To(MatchAllKeys(Keys{
				"status":    BeNumerically("==", http.StatusCreated),
				"message":   Equal("The resource has been successfully created on the server."),
				"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				"data": MatchAllKeys(Keys{
					"id":        extmatchers.BeValidUUID(),
					"username":  Equal("tester"),
					"email":     Equal("tester@internal.com"),
					"createdAt": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
					"createdBy": Equal("system"),
					"updatedAt": BeNil(),
					"updatedBy": BeNil(),
				}),
			}))
		})
	})

	Context("with invalid registration data", func() {
		It("should return validation errors", func() {
			response := httpexpecter.POST("/auth/register").WithJSON(map[string]string{
				"username": "",
				"email":    "",
				"password": "",
			}).Expect()

			response.Status(http.StatusBadRequest)
			Expect(response.JSON().Object().Raw()).To(MatchAllKeys(Keys{
				"status":    BeNumerically("==", http.StatusBadRequest),
				"message":   Equal("The request contains invalid parameters or is malformed."),
				"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				"data": MatchAllKeys(Keys{
					"RegisterFormData.Username": Equal("Username is a required field"),
					"RegisterFormData.Password": Equal("Password is a required field"),
					"RegisterFormData.Email":    Equal("Email is a required field"),
				}),
			}))
		})
	})

	Context("create a new user by using another user", func() {
		It("should return createdBy = tester", func() {
			response := httpexpecter.POST("/auth/register").WithJSON(map[string]string{
				"username": "tester",
				"email":    "tester@internal.com",
				"password": "Keep!T5ecret",
			}).Expect()

			response.Status(http.StatusCreated)

			loginResponse := httpexpecter.POST("/auth/login").WithJSON(map[string]string{
				"username": "tester",
				"password": "Keep!T5ecret",
			}).Expect()

			session_id := loginResponse.Cookie("session-id").Value()

			registerResponse := httpexpecter.POST("/auth/register").
				WithCookie("session_id", session_id.Raw()).
				WithJSON(map[string]string{
					"username": "tester_02",
					"email":    "tester_02@internal.com",
					"password": "Keep!T5ecret",
				}).
				Expect()

			registerResponse.Status(http.StatusCreated)
			Expect(registerResponse.JSON().Object().Raw()).To(MatchAllKeys(Keys{
				"status":    BeNumerically("==", http.StatusCreated),
				"message":   Equal("The resource has been successfully created on the server."),
				"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				"data": MatchAllKeys(Keys{
					"id":        extmatchers.BeValidUUID(),
					"username":  Equal("tester_02"),
					"email":     Equal("tester_02@internal.com"),
					"createdAt": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
					"createdBy": Equal("tester"),
					"updatedAt": BeNil(),
					"updatedBy": BeNil(),
				}),
			}))
		})
	})
})
