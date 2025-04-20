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

func (s *HealthzRouteTestSuite) SetupSuite() {
	s.StartPostgresContainer()
}

func (s *HealthzRouteTestSuite) SetupTest() {
	s.RestorePostgresContainer()
}

func (s *HealthzRouteTestSuite) TearDownSuite() {
	s.TerminateTestPostgres()
}

func (s *HealthzRouteTestSuite) TestHealthCheckRoute() {
	di := s.SetupTestDependencyContainer(s.T())
	router := s.DefaultWebRouter(
		s.T(),
		s.WithRoute(healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)),
	)

	s.True(router.Match(chi.NewRouteContext(), http.MethodGet, "/healthz"))
}

func (s *HealthzRouteTestSuite) TestHealthCheckRouteServeHTTP() {
	di := s.SetupTestDependencyContainer(s.T())
	healthzRoute := healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)
	router := s.DefaultWebRouter(s.T(), s.WithRoute(healthzRoute))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), nil)

	router.ServeHTTP(w, r)

	var response response.Response[map[string]string]

	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &response))

	s.Equal(http.StatusOK, w.Code)
	s.Equal(http.StatusOK, response.Status)
	s.Equal("Your request was successfully completed.", response.Message)
	s.Equal("available", response.Data["database"])
	s.WithinDuration(time.Now(), response.Timestamp, time.Second)
}

func (s *HealthzRouteTestSuite) TestHealthCheckRouteServeHTTPV2() {
	di := s.SetupTestDependencyContainer(s.T())

	healthzRoute := healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)
	router := s.DefaultWebRouter(s.T(), s.WithRoute(healthzRoute))

	// Simulate a database connection error
	s.TerminateTestPostgres()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), nil)

	router.ServeHTTP(w, r)

	var response response.Response[map[string]string]

	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &response))

	s.Equal(http.StatusServiceUnavailable, w.Code)
	s.Equal(http.StatusServiceUnavailable, response.Status)
	s.Equal("The service is currently unavailable; please try again later.", response.Message)
	s.Equal("unavailable", response.Data["database"])
	s.WithinDuration(time.Now(), response.Timestamp, time.Second)
}
