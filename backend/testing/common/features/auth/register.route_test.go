package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal.snowdrop/common/core/trans"
	"internal.snowdrop/common/core/web"
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	mockconfig "internal.snowdrop/testing/mocks/internal.snowdrop/common/core/config"
	mockusermgt "internal.snowdrop/testing/mocks/internal.snowdrop/common/features/usermgt"
	"internal.snowdrop/testing/testutils"
)

func TestRegisterRoute(t *testing.T) {
	t.Parallel()

	registerRoute := auth.NewRegisterRoute(auth.RegisterRouteParams{
		Validator:   web.NewValidator(),
		UserService: mockusermgt.NewMockUserService(t),
	})

	router := web.NewRouter(web.RouteParams{
		HTTPRoutes: []web.HTTPHandler{registerRoute},
	})

	assert.True(t, router.Match(chi.NewRouteContext(), http.MethodPost, "/auth/register"))
}

func TestRegisterRouteServeHTTP(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	})

	config := mockconfig.NewMockManager(t)
	config.EXPECT().GetEmbedResourceFolder().Return(testutils.GetResourceFS())
	bundle, err := trans.NewI18nBundle(config)
	require.NoError(t, err)

	userRepository := usermgt.NewUserRepository(db)
	userService := usermgt.NewUserService(userRepository)
	registerRoute := auth.NewRegisterRoute(auth.RegisterRouteParams{
		Validator:   web.NewValidator(),
		UserService: userService,
	})
	router := web.NewRouter(web.RouteParams{
		I18nBundle: bundle,
		HTTPRoutes: []web.HTTPHandler{registerRoute},
	})

	mock.MatchExpectationsInOrder(true)
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO public.users`).
		WithArgs(sqlmock.AnyArg(), "admin", "admin@internal.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "created_by"}).AddRow(time.Now(), "system"))
	mock.ExpectCommit()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(registerRoute.Method(), registerRoute.Path(), bytes.NewBufferString(`
	{
		"username": "admin",
		"email": "admin@internal.com",
		"password": "Keep!T5ecret"
	}`))

	router.ServeHTTP(w, r)

	var result web.Response[usermgt.UserDTO]

	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, http.StatusCreated, result.Status)
	assert.Equal(t, "The resource has been successfully created on the server.", result.Message)
	require.NoError(t, uuid.Validate(result.Data.ID))
	assert.Equal(t, "admin", result.Data.Username)
	assert.Equal(t, "admin@internal.com", result.Data.Email)
	assert.WithinDuration(t, time.Now(), result.Data.CreatedAt, time.Second)
	assert.Equal(t, "system", result.Data.CreatedBy)
	assert.Nil(t, result.Data.UpdatedAt)
	assert.Nil(t, result.Data.UpdatedBy)
	assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)
}
