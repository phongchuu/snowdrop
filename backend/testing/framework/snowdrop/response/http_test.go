package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/validation"
	mocksnowdrop "internal.snowdrop/testing/mocks/internal.snowdrop/framework"
	"internal.snowdrop/testing/testutils"
)

var _ = Describe("ResponseBuilder", func() {
	var (
		w *httptest.ResponseRecorder
		r *http.Request
	)

	BeforeEach(func() {
		config := mocksnowdrop.NewMockConfigManager(GinkgoT())
		config.EXPECT().GetEmbedResourceFolder().Return(testutils.GetResourceFS())

		bundle, err := translation.NewI18nBundle(config)
		Expect(err).NotTo(HaveOccurred())

		validator, err := validation.NewValidator()
		Expect(err).NotTo(HaveOccurred())

		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		r = translation.WithLocalizer(r, i18n.NewLocalizer(bundle))
		r = validation.WithUniversalTranslator(r, validator.UniversalTranslator.GetFallback())
	})

	It("should build a valid response", func() {
		response.NewBuilder(w, r).
			Status(http.StatusOK).
			Data(map[string]string{"message": "OK"}).
			JSON()

		var result response.Response[map[string]string]
		err := json.Unmarshal(w.Body.Bytes(), &result)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Status).To(Equal(http.StatusOK))
		Expect(w.Header().Get("X-Content-Type-Options")).To(Equal("nosniff"))
		Expect(result.Data).To(Equal(map[string]string{"message": "OK"}))
		Expect(result.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
	})
})
