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
	"internal.snowdrop/common/features/auth"
	"internal.snowdrop/common/features/usermgt"
	"internal.snowdrop/framework/web"
	"internal.snowdrop/testing/testutils"
)

func TestRegisterRoute(t *testing.T) {
	t.Parallel()

	di := testutils.SetupTestDependencyContainer(t)
	router := testutils.DefaultWebRouter(
		t,
		testutils.WithDI(di),
		testutils.WithRoute(auth.NewRegisterRoute(auth.RegisterRouteParams{
			Validator:   di.Validator,
			UserService: di.MockUserService,
		})),
	)

	assert.True(t, router.Match(chi.NewRouteContext(), http.MethodPost, "/auth/register"))
}

func TestRegisterRouteServeHTTP(t *testing.T) {
	t.Parallel()

	di := testutils.SetupTestDependencyContainer(t)

	userRepository := usermgt.NewUserRepository(usermgt.UserRepositoryParams{
		TransactionManager: di.TransactionManager,
	})
	userService := usermgt.NewUserService(userRepository, di.NoopTracer)
	authRegisterRoute := auth.NewRegisterRoute(auth.RegisterRouteParams{
		Validator:   di.Validator,
		UserService: userService,
	})
	router := testutils.DefaultWebRouter(
		t,
		testutils.WithDI(di),
		testutils.WithRoute(authRegisterRoute),
	)

	di.SQLMock.ExpectBegin()
	di.SQLMock.ExpectExec(`SELECT set_config`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	di.SQLMock.ExpectQuery(`INSERT INTO sessions`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "created_at", "last_accessed_at", "expires_at", "data"}).
				AddRow("random-id", uuid.Nil, time.Time{}, time.Time{}, time.Time{}, ""),
		)
	di.SQLMock.ExpectCommit()

	di.SQLMock.ExpectBegin()
	di.SQLMock.ExpectExec(`SELECT set_config`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	di.SQLMock.ExpectQuery(`INSERT INTO public.users`).
		WithArgs(sqlmock.AnyArg(), "admin", "admin@internal.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "created_by"}).AddRow(time.Now(), "system"))
	di.SQLMock.ExpectCommit()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`
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
	assert.NotEqual(t, result.Data.ID, uuid.Nil)
	assert.Equal(t, "admin", result.Data.Username)
	assert.Equal(t, "admin@internal.com", result.Data.Email)
	assert.WithinDuration(t, time.Now(), result.Data.CreatedAt, time.Second)
	assert.Equal(t, "system", result.Data.CreatedBy)
	assert.Nil(t, result.Data.UpdatedAt)
	assert.Nil(t, result.Data.UpdatedBy)
	assert.WithinDuration(t, time.Now(), result.Timestamp, time.Second)
}
