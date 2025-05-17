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
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/testing/testutils"
	"internal.snowdrop/testing/testutils/extmatchers"
)

var _ = Describe("[RegisterRoute]", Label("integration"), Serial, func() {
	var (
		server       *httptest.Server
		httpexpecter *httpexpect.Expect
	)

	BeforeEach(func() {
		di := testutils.IntegrationTestSetup(GinkgoTB())
		userRepository := usermgt.NewUserRepository(usermgt.UserRepositoryParams{
			TransactionManager: di.TransactionManager,
		})
		userService := usermgt.NewUserService(userRepository, di.NoopTracer)
		authRegisterRoute := auth.NewRegisterRoute(auth.RegisterRouteParams{
			Validator:   di.Validator,
			UserService: userService,
		})
		router := testutils.NewRouter(di, []snowdrop.HTTPHandler{
			authRegisterRoute,
		})
		server = httptest.NewServer(router)
		httpexpecter = httpexpect.Default(GinkgoT(), server.URL)
	})

	AfterEach(func() {
		server.Close()
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
})
