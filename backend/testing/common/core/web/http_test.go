package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal.snowdrop/common/core/web"
)

func TestResponseBuilder(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	web.NewResponseBuilder(w, r).
		Status(http.StatusOK).
		Data(map[string]string{"message": "OK"}).
		JSON()

	var result web.Response[map[string]string]
	err := json.Unmarshal(w.Body.Bytes(), &result)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Status)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Exactly(t, map[string]string{"message": "OK"}, result.Data)
	assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)
}

func BenchmarkResponseBuilder(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		responseBuilder := web.NewResponseBuilder(w, r)
		responseBuilder.Status(http.StatusOK).JSON()
	}
}
