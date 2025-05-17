package healthz_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gavv/httpexpect/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/spf13/cast"
	"internal.snowdrop/common/features/healthz"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/testing/testutils"
)

var _ = Describe("[HealthCheckRoute]", Label("integration"), Serial, func() {
	var (
		di           testutils.TestContainer
		server       *httptest.Server
		httpexpecter *httpexpect.Expect
	)

	BeforeEach(func() {
		di = testutils.IntegrationTestSetup(GinkgoTB())
		healthzRoute := healthz.NewHealthCheckRoute(di.NoopLogger, di.DB)
		router := testutils.NewRouter(di, []snowdrop.HTTPHandler{
			healthzRoute,
		})
		server = httptest.NewServer(router)
		httpexpecter = httpexpect.Default(GinkgoT(), server.URL)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe(".ServeHTTP", func() {
		Context("there is no wrong", func() {
			It("should no error", func() {
				response := httpexpecter.GET("/healthz").Expect()
				response.Status(http.StatusOK)
				Expect(response.JSON().Object().Raw()).To(gstruct.MatchAllKeys(gstruct.Keys{
					"status":  BeNumerically("==", http.StatusOK),
					"message": Equal("Your request was successfully completed."),
					"data": gstruct.MatchAllKeys(gstruct.Keys{
						"database": Equal("available"),
					}),
					"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				}))
			})
		})

		Context("When database is down", func() {
			It("should return an error", func(ctx SpecContext) {
				Expect(di.PgContainer.Terminate(ctx)).To(Succeed())

				response := httpexpecter.GET("/healthz").Expect()
				response.Status(http.StatusServiceUnavailable)
				Expect(response.JSON().Object().Raw()).To(gstruct.MatchAllKeys(gstruct.Keys{
					"status":  BeNumerically("==", http.StatusServiceUnavailable),
					"message": Equal("The service is currently unavailable; please try again later."),
					"data": gstruct.MatchAllKeys(gstruct.Keys{
						"database": Equal("unavailable"),
					}),
					"timestamp": WithTransform(cast.ToTimeE, BeTemporally("~", time.Now(), time.Second)),
				}))
			})
		})
	})
})
