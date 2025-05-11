package healthz_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"internal.snowdrop/common/features/healthz"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/testing/testutils"
)

var _ = Describe("[HealthCheckRoute]", Label("integration"), Serial, func() {
	var (
		di           testutils.TestContainer
		router       *chi.Mux
		healthzRoute snowdrop.HTTPHandler
	)

	BeforeEach(func() {
		di = testutils.IntegrationTestSetup(GinkgoTB())
		healthzRoute = healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)
		router = testutils.NewRouter(di, []snowdrop.HTTPHandler{
			healthzRoute,
		})
	})

	Context("When verifying route configuration", func() {
		It("should have a endpoint: GET /healthz", func() {
			isMatched := router.Match(chi.NewRouteContext(), http.MethodGet, "/healthz")
			Expect(isMatched).To(BeTrue())
		})
	})

	Describe(".ServeHTTP", func() {
		Context("there is no wrong", func() {
			It("should no error", func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), http.NoBody)

				router.ServeHTTP(w, r)

				var result response.Response[map[string]string]
				Expect(json.Unmarshal(w.Body.Bytes(), &result)).To(Succeed())

				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(result.Status).To(Equal(http.StatusOK))
				Expect(result.Message).To(Equal("Your request was successfully completed."))
				Expect(result.Data["database"]).To(Equal("available"))
				Expect(result.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
			})
		})

		Context("When database is down", func() {
			It("should return an error", func(ctx SpecContext) {
				Expect(di.PgContainer.Terminate(ctx)).To(Succeed())

				w := httptest.NewRecorder()
				r := httptest.NewRequest(healthzRoute.Method(), healthzRoute.Path(), nil)

				router.ServeHTTP(w, r)

				var result response.Response[map[string]string]
				Expect(json.Unmarshal(w.Body.Bytes(), &result)).To(Succeed())

				Expect(w.Code).To(Equal(http.StatusServiceUnavailable))
				Expect(result.Status).To(Equal(http.StatusServiceUnavailable))
				Expect(result.Message).To(Equal("The service is currently unavailable; please try again later."))
				Expect(result.Data["database"]).To(Equal("unavailable"))
				Expect(result.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
			})
		})
	})
})
