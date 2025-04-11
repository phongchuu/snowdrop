package testutils

import (
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/text/language"
	"internal.snowdrop/common/core"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/trans"
	"internal.snowdrop/framework/web"
	mockusermgt "internal.snowdrop/testing/mocks/internal.snowdrop/common/features/usermgt"
	mocksnowdrop "internal.snowdrop/testing/mocks/internal.snowdrop/framework"
)

type TestContainer struct {
	NoopLogger                    *slog.Logger
	MockConfigManager             *mocksnowdrop.MockConfigManager
	I18nBundle                    *i18n.Bundle
	Validator                     *validator.Validate
	UniversalTranslator           *ut.UniversalTranslator
	I18nMiddleware                trans.I18nMiddleware
	UniversalTranslatorMiddleware web.UniversalTranslatorMiddleware
	RecovererMiddleware           web.RecovererMiddleware
	NoopTracer                    trace.Tracer
	TransactionManager            snowdrop.TransactionManager
	SessionManager                *session.Manager
	MockUserService               *mockusermgt.MockUserService
	SQLMock                       sqlmock.Sqlmock
}

type DefaultWebRouterOptions struct {
	di     *TestContainer
	routes []snowdrop.HTTPHandler
}

func getWorkspaceDir() string {
	cmd := exec.Command("go", "env", "GOWORK")
	output, _ := cmd.Output()

	return filepath.Dir(string(output))
}

// GetResourceFS creates a virtual file system (fs.FS).
func GetResourceFS() fs.FS {
	dirPath := filepath.Join(getWorkspaceDir(), "console", "resources", "trans")
	files, _ := os.ReadDir(dirPath)

	virtualFileSys := fstest.MapFS{}

	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())
		virtualPath := filepath.Join("resources", "trans", file.Name())
		data, _ := os.ReadFile(filePath)

		virtualFileSys[virtualPath] = &fstest.MapFile{Data: data}
	}

	return virtualFileSys
}

func SetupTestDependencyContainer(t testing.TB) TestContainer {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	})

	noopLogger := log.NewNoopLogger()

	configManager := mocksnowdrop.NewMockConfigManager(t)
	configManager.EXPECT().GetEmbedResourceFolder().Return(GetResourceFS())
	configManager.EXPECT().GetSupportedLanguages().Return([]language.Tag{
		language.English,
		language.Vietnamese,
	})

	i18nBundle, err := trans.NewI18nBundle(configManager)
	require.NoError(t, err)

	validationModule, err := web.NewValidator()
	require.NoError(t, err)

	getPreferredUserLanguageFn := core.NewGetPreferredUserLanguageFn(configManager)

	transactionManager := database.NewTransactionManager(db)

	sessionManager := session.NewManager(session.ManagerParams{
		TransactionManager: transactionManager,
		Repository:         session.NewPostgresRepository(transactionManager),
	})

	return TestContainer{
		SQLMock:             mock,
		NoopLogger:          noopLogger,
		MockConfigManager:   configManager,
		I18nBundle:          i18nBundle,
		Validator:           validationModule.Validator,
		UniversalTranslator: validationModule.Uni,
		I18nMiddleware: trans.NewI18nMiddleware(
			i18nBundle,
			getPreferredUserLanguageFn,
		),
		UniversalTranslatorMiddleware: web.NewUniversalTranslatorMiddleware(
			validationModule.Uni,
			getPreferredUserLanguageFn,
		),
		RecovererMiddleware: web.NewRecovererMiddleware(noopLogger),
		NoopTracer:          noop.NewTracerProvider().Tracer("noop-tracer"),
		TransactionManager:  transactionManager,
		SessionManager:      sessionManager,
		MockUserService:     mockusermgt.NewMockUserService(t),
	}
}

func WithDI(di TestContainer) func(*DefaultWebRouterOptions) {
	return func(dwro *DefaultWebRouterOptions) {
		dwro.di = &di
	}
}

func WithRoute(route snowdrop.HTTPHandler) func(*DefaultWebRouterOptions) {
	return func(dwro *DefaultWebRouterOptions) {
		dwro.routes = append(dwro.routes, route)
	}
}

func DefaultWebRouter(t testing.TB, optionFns ...func(*DefaultWebRouterOptions)) *chi.Mux {
	options := DefaultWebRouterOptions{}

	for i := range optionFns {
		optionFns[i](&options)
	}

	if options.di == nil {
		options.di = lo.ToPtr(SetupTestDependencyContainer(t))
	}

	return web.NewRouter(web.RouteParams{
		Logger:                        options.di.NoopLogger,
		I18nBundle:                    options.di.I18nBundle,
		I18nMiddleware:                options.di.I18nMiddleware,
		SessionManager:                options.di.SessionManager,
		UniversalTranslatorMiddleware: options.di.UniversalTranslatorMiddleware,
		RecovererMiddleware:           options.di.RecovererMiddleware,
		Tracer:                        options.di.NoopTracer,
		HTTPRoutes:                    options.routes,
	})
}
