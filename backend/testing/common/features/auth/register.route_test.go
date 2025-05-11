package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/testing/testutils"
)

var _ = Describe("[RegisterRoute]", Label("integration"), Serial, func() {
	var (
		di                testutils.TestContainer
		router            *chi.Mux
		authRegisterRoute snowdrop.HTTPHandler
	)

	BeforeEach(func() {
		di = testutils.IntegrationTestSetup(GinkgoTB())

		userRepository := usermgt.NewUserRepository(usermgt.UserRepositoryParams{
			TransactionManager: di.TransactionManager,
		})
		userService := usermgt.NewUserService(userRepository, di.NoopTracer)
		authRegisterRoute = auth.NewRegisterRoute(auth.RegisterRouteParams{
			Validator:   di.Validator,
			UserService: userService,
		})

		router = testutils.NewRouter(di, []snowdrop.HTTPHandler{
			authRegisterRoute,
		})
	})

	Context("When verifying route configuration", func() {
		It("should have a endpoint: POST /auth/register", func() {
			matched := router.Match(
				chi.NewRouteContext(),
				http.MethodPost,
				"/auth/register",
			)

			Expect(matched).To(BeTrue())
		})
	})

	Context("with valid registration data", func() {
		It("should successfully create a new user", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, authRegisterRoute.Path(), bytes.NewBufferString(`{
                "username": "tester",
                "email": "tester@internal.com",
                "password": "Keep!T5ecret"
            }`))

			router.ServeHTTP(w, r)

			var result response.Response[usermgt.UserDTO]
			Expect(json.NewDecoder(w.Body).Decode(&result)).To(Succeed())

			// Verify HTTP response
			Expect(w.Code).To(Equal(http.StatusCreated))

			// Verify response content
			Expect(result.Status).To(Equal(http.StatusCreated))
			Expect(result.Message).To(Equal("The resource has been successfully created on the server."))

			// Verify user data
			Expect(result.Data.ID).NotTo(Equal(uuid.Nil))
			Expect(result.Data.Username).To(Equal("tester"))
			Expect(result.Data.Email).To(Equal("tester@internal.com"))
			Expect(result.Data.CreatedAt).To(BeTemporally("~", time.Now(), time.Second))
			Expect(result.Data.CreatedBy).To(Equal("system"))
			Expect(result.Data.UpdatedAt).To(BeNil())
			Expect(result.Data.UpdatedBy).To(BeNil())

			// Verify timestamp
			Expect(result.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
		})
	})

	Context("with invalid registration data", func() {
		It("should return validation errors", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, authRegisterRoute.Path(), bytes.NewBufferString(`{
                "username": "",
                "email": "",
                "password": ""
            }`))

			router.ServeHTTP(w, r)

			var result response.Response[validator.ValidationErrorsTranslations]
			Expect(json.NewDecoder(w.Body).Decode(&result)).To(Succeed())

			// Verify HTTP response
			Expect(w.Code).To(Equal(http.StatusBadRequest))

			// Verify response content
			Expect(result.Status).To(Equal(http.StatusBadRequest))
			Expect(result.Message).To(Equal("The request contains invalid parameters or is malformed."))

			// Verify validation errors
			Expect(result.Data).To(HaveLen(3))
			Expect(result.Data["RegisterFormData.Username"]).To(Equal("Username is a required field"))
			Expect(result.Data["RegisterFormData.Password"]).To(Equal("Password is a required field"))
			Expect(result.Data["RegisterFormData.Email"]).To(Equal("Email is a required field"))

			// Verify timestamp
			Expect(result.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
		})
	})
})
