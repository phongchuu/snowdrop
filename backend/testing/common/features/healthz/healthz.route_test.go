package healthz_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"internal.snowdrop/common/features/healthz"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/testing/testutils"
)

type HealthzRouteTestSuite struct {
	testutils.ExtendedTestSuite
}

//nolint:paralleltest // This test uses a database container, so it cannot run in parallel.
func TestHealthzRouteTestSuite(t *testing.T) {
	suite.Run(t, new(HealthzRouteTestSuite))
}

func (s *HealthzRouteTestSuite) SetupTest() {
	s.StartPostgresContainer()
}

func (s *HealthzRouteTestSuite) TearDownTest() {
	s.RestorePostgresContainer()
}

func (s *HealthzRouteTestSuite) TestHealthCheckRoute() {
	di := s.SetupTestDependencyContainer()
	router := s.DefaultWebRouter(
		s.WithRoute(healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)),
	)

	s.True(router.Match(chi.NewRouteContext(), http.MethodGet, "/healthz"))
}

func (s *HealthzRouteTestSuite) TestHealthCheckRouteServeHTTP() {
	di := s.SetupTestDependencyContainer()
	healthzRoute := healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)
	router := s.DefaultWebRouter(s.WithRoute(healthzRoute))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), nil)

	router.ServeHTTP(w, r)

	var result response.Response[map[string]string]

	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))

	s.Equal(http.StatusOK, w.Code)
	s.Equal(http.StatusOK, result.Status)
	s.Equal("Your request was successfully completed.", result.Message)
	s.Equal("available", result.Data["database"])
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}

func (s *HealthzRouteTestSuite) TestHealthCheckRouteServeHTTPV2() {
	di := s.SetupTestDependencyContainer()

	healthzRoute := healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)
	router := s.DefaultWebRouter(s.WithRoute(healthzRoute))

	// Simulate a database connection error
	s.StopPostgresContainer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), nil)

	router.ServeHTTP(w, r)

	var result response.Response[map[string]string]

	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &result))

	s.Equal(http.StatusServiceUnavailable, w.Code)
	s.Equal(http.StatusServiceUnavailable, result.Status)
	s.Equal("The service is currently unavailable; please try again later.", result.Message)
	s.Equal("unavailable", result.Data["database"])
	s.WithinDuration(time.Now(), result.Timestamp, time.Second)
}
