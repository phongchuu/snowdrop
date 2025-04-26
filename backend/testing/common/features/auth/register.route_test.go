package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/testing/testutils"
)

type RegisterRouteTestSuite struct {
	testutils.ExtendedTestSuite
}

//nolint:paralleltest // This test uses a database container, so it cannot run in parallel.
func TestRegisterRouteSuite(t *testing.T) {
	suite.Run(t, new(RegisterRouteTestSuite))
}

func (s *RegisterRouteTestSuite) SetupTest() {
	s.StartPostgresContainer()
}

func (s *RegisterRouteTestSuite) TearDownTest() {
	s.RestorePostgresContainer()
}

func (s *RegisterRouteTestSuite) TestRegisterRoute() {
	di := s.SetupTestDependencyContainer(s.T())
	router := s.DefaultWebRouter(
		s.T(),
		s.WithDI(di),
		s.WithRoute(auth.NewRegisterRoute(auth.RegisterRouteParams{
			Validator:   di.Validator,
			UserService: di.MockUserService,
		})),
	)

	s.True(router.Match(chi.NewRouteContext(), http.MethodPost, "/auth/register"))
}

func (s *RegisterRouteTestSuite) TestRegisterRouteServeHTTP() {
	di := s.SetupTestDependencyContainer(s.T())

	userRepository := usermgt.NewUserRepository(usermgt.UserRepositoryParams{
		TransactionManager: di.TransactionManager,
	})
	userService := usermgt.NewUserService(userRepository, di.NoopTracer)
	authRegisterRoute := auth.NewRegisterRoute(auth.RegisterRouteParams{
		Validator:   di.Validator,
		UserService: userService,
	})
	router := s.DefaultWebRouter(
		s.T(),
		s.WithDI(di),
		s.WithRoute(authRegisterRoute),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, authRegisterRoute.Path(), bytes.NewBufferString(`
	{
		"username": "tester",
		"email": "tester@internal.com",
		"password": "Keep!T5ecret"
	}`))

	router.ServeHTTP(w, r)

	var result response.Response[usermgt.UserDTO]

	s.Require().NoError(json.NewDecoder(w.Body).Decode(&result))
	s.Equal(http.StatusCreated, w.Code)
	s.Equal(http.StatusCreated, result.Status)
	s.Equal("The resource has been successfully created on the server.", result.Message)
	s.NotEqual(result.Data.ID, uuid.Nil)
	s.Equal("tester", result.Data.Username)
	s.Equal("tester@internal.com", result.Data.Email)
	s.WithinDuration(time.Now(), result.Data.CreatedAt, time.Second)
	s.Equal("system", result.Data.CreatedBy)
	s.Nil(result.Data.UpdatedAt)
	s.Nil(result.Data.UpdatedBy)
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}

func (s *RegisterRouteTestSuite) TestRegisterRouteHTTPValidationErrors() {
	di := s.SetupTestDependencyContainer(s.T())

	userRepository := usermgt.NewUserRepository(usermgt.UserRepositoryParams{
		TransactionManager: di.TransactionManager,
	})
	userService := usermgt.NewUserService(userRepository, di.NoopTracer)
	authRegisterRoute := auth.NewRegisterRoute(auth.RegisterRouteParams{
		Validator:   di.Validator,
		UserService: userService,
	})
	router := s.DefaultWebRouter(
		s.T(),
		s.WithDI(di),
		s.WithRoute(authRegisterRoute),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, authRegisterRoute.Path(), bytes.NewBufferString(`
	{
		"username": "",
		"email": "",
		"password": ""
	}`))

	router.ServeHTTP(w, r)

	var result response.Response[validator.ValidationErrorsTranslations]

	s.Require().NoError(json.NewDecoder(w.Body).Decode(&result))
	s.Equal(http.StatusBadRequest, w.Code)
	s.Equal(http.StatusBadRequest, result.Status)
	s.Equal("The request contains invalid parameters or is malformed.", result.Message)
	s.Len(result.Data, 3)
	s.Equal("Username is a required field", result.Data["RegisterFormData.Username"])
	s.Equal("Password is a required field", result.Data["RegisterFormData.Password"])
	s.Equal("Email is a required field", result.Data["RegisterFormData.Email"])
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}
