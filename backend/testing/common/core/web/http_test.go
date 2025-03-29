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
	"internal.snowdrop/common/core/trans"
	"internal.snowdrop/common/core/web"
	mockconfig "internal.snowdrop/testing/mocks/internal.snowdrop/common/core/config"
	"internal.snowdrop/testing/testutils"
)

func TestResponseBuilder(t *testing.T) {
	config := mockconfig.NewMockManager(t)
	config.EXPECT().GetEmbedResourceFolder().Return(testutils.GetResourceFS())
	bundle, err := trans.NewI18nBundle(config)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := trans.WithLocalizer(httptest.NewRequest(http.MethodGet, "/", nil), i18n.NewLocalizer(bundle))

	web.NewResponseBuilder(w, r).
		Status(http.StatusOK).
		Data(map[string]string{"message": "OK"}).
		JSON()

	var result web.Response[map[string]string]
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.Status)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Exactly(t, map[string]string{"message": "OK"}, result.Data)
	assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)
}
