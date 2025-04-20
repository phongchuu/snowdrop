package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal.snowdrop/framework/response"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/validation"
	mocksnowdrop "internal.snowdrop/testing/mocks/internal.snowdrop/framework"
	"internal.snowdrop/testing/testutils"
)

func TestResponseBuilder(t *testing.T) {
	t.Parallel()

	config := mocksnowdrop.NewMockConfigManager(t)
	config.EXPECT().GetEmbedResourceFolder().Return(testutils.GetResourceFS())
	bundle, err := translation.NewI18nBundle(config)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := translation.WithLocalizer(
		httptest.NewRequest(http.MethodGet, "/", nil),
		i18n.NewLocalizer(bundle),
	)

	validator, err := validation.NewValidator()
	require.NoError(t, err)

	r = validation.WithUniversalTranslator(
		r,
		validator.UniversalTranslator.GetFallback(),
	)

	response.NewBuilder(w, r).
		Status(http.StatusOK).
		Data(map[string]string{"message": "OK"}).
		JSON()

	var result response.Response[map[string]string]
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Status)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Exactly(t, map[string]string{"message": "OK"}, result.Data)
	assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)
}
