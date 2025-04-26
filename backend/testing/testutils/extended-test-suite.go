package testutils

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/go-chi/chi/v5"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx/fxtest"
	"golang.org/x/text/language"
	"internal.snowdrop/common/core"
	snowdrop "internal.snowdrop/framework"
	"internal.snowdrop/framework/database"
	"internal.snowdrop/framework/log"
	"internal.snowdrop/framework/session"
	"internal.snowdrop/framework/translation"
	"internal.snowdrop/framework/validation"
	"internal.snowdrop/framework/web"
	mockusermgt "internal.snowdrop/testing/mocks/internal.snowdrop/common/features/usermgt"
	mocksnowdrop "internal.snowdrop/testing/mocks/internal.snowdrop/framework"
)

type TestContainer struct {
	NoopLogger                 *slog.Logger
	MockConfigManager          *mocksnowdrop.MockConfigManager
	I18nBundle                 *i18n.Bundle
	Validator                  *validator.Validate
	UniversalTranslator        *ut.UniversalTranslator
	NoopTracer                 trace.Tracer
	TransactionManager         snowdrop.TransactionManager
	SessionManager             *session.Manager
	MockUserService            *mockusermgt.MockUserService
	GetPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn
	DB                         *sql.DB
}

type DefaultWebRouterOptions struct {
	di     *TestContainer
	routes []snowdrop.HTTPHandler
}

type ExtendedTestSuite struct {
	suite.Suite
	pgContainer      *postgres.PostgresContainer
	connectionString string
}

// StartPostgresContainer starts a PostgreSQL container for tests.
// The container is reusable and a snapshot is created after startup.
func (s *ExtendedTestSuite) StartPostgresContainer() {
	s.T().Helper()

	ctx := s.T().Context()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("snowdrop"),
		postgres.WithUsername("tester"),
		postgres.WithPassword("Keep!t5ecret"),
		postgres.BasicWaitStrategies(),
		testcontainers.WithReuseByName("e2e-postgresql"),
	)
	s.Require().NoError(err, "Failed to start test postgres container")

	err = pgContainer.Snapshot(ctx)
	s.Require().NoError(err, "Failed to create snapshot of test postgres container")

	s.pgContainer = pgContainer
	s.connectionString = pgContainer.MustConnectionString(ctx, "sslmode=disable")
}

// RestorePostgresContainer restores the PostgreSQL container to its initial state.
func (s *ExtendedTestSuite) RestorePostgresContainer() {
	s.T().Helper()

	if s.pgContainer.IsRunning() {
		s.Require().NoError(s.pgContainer.Restore(s.T().Context()))
	}
}

// StopPostgresContainer stops the PostgreSQL container.
func (s *ExtendedTestSuite) StopPostgresContainer() {
	s.T().Helper()

	if s.pgContainer.IsRunning() {
		err := s.pgContainer.Terminate(s.T().Context())
		s.Require().NoError(err)
	}
}

func (s *ExtendedTestSuite) SetupTestDependencyContainer(tb testing.TB) TestContainer {
	tb.Helper()

	// This is used for database migration
	tb.Setenv("APP_DEFAULT_ADMIN_PASSWORD", "Keep!t5ecret")

	fxLifecycle := fxtest.NewLifecycle(tb)
	noopLogger := log.NewNoopLogger()

	// ConfigManager
	configManager := mocksnowdrop.NewMockConfigManager(tb)
	configManager.EXPECT().GetEmbedResourceFolder().Return(GetResourceFS())
	configManager.EXPECT().GetSupportedLanguages().Return([]language.Tag{
		language.English,
		language.Vietnamese,
	})
	configManager.EXPECT().GetMaxRequestSize().Return(int64(1 << 20)).Maybe()
	configManager.EXPECT().
		GetDatabaseURL().
		Return(s.connectionString)

	// I18nBundle
	i18nBundle, err := translation.NewI18nBundle(configManager)
	require.NoError(tb, err)

	// Validator
	validationModule, err := validation.NewValidator()
	require.NoError(tb, err)

	// GetPreferredUserLanguageFn
	getPreferredUserLanguageFn := core.NewGetPreferredUserLanguageFn(configManager)

	// Database
	sqlDB, err := database.NewDatabase(fxLifecycle, noopLogger, configManager)
	require.NoError(tb, err)
	database.ExecuteDatabaseUpgrade(fxLifecycle, noopLogger, sqlDB, configManager)

	// Start the lifecycle
	fxLifecycle.RequireStart()
	tb.Cleanup(func() {
		fxLifecycle.RequireStop()
	})

	// TransactionManager
	transactionManager := database.NewTransactionManager(sqlDB)

	// SessionManager
	sessionManager := session.NewManager(session.ManagerParams{
		TransactionManager: transactionManager,
		Repository:         session.NewPostgresRepository(transactionManager),
	})

	return TestContainer{
		GetPreferredUserLanguageFn: getPreferredUserLanguageFn,
		NoopLogger:                 noopLogger,
		MockConfigManager:          configManager,
		I18nBundle:                 i18nBundle,
		Validator:                  validationModule.Validator,
		UniversalTranslator:        validationModule.UniversalTranslator,
		NoopTracer:                 noop.NewTracerProvider().Tracer("noop-tracer"),
		TransactionManager:         transactionManager,
		SessionManager:             sessionManager,
		MockUserService:            mockusermgt.NewMockUserService(tb),
		DB:                         sqlDB,
	}
}

func (s *ExtendedTestSuite) WithDI(di TestContainer) func(*DefaultWebRouterOptions) {
	return func(dwro *DefaultWebRouterOptions) {
		dwro.di = &di
	}
}

func (s *ExtendedTestSuite) WithRoute(route snowdrop.HTTPHandler) func(*DefaultWebRouterOptions) {
	return func(dwro *DefaultWebRouterOptions) {
		dwro.routes = append(dwro.routes, route)
	}
}

func (s *ExtendedTestSuite) DefaultWebRouter(
	t testing.TB,
	optionFns ...func(*DefaultWebRouterOptions),
) *chi.Mux {
	options := DefaultWebRouterOptions{}

	for i := range optionFns {
		optionFns[i](&options)
	}

	if options.di == nil {
		options.di = lo.ToPtr(s.SetupTestDependencyContainer(t))
	}

	return web.NewRouter(web.RouteParams{
		Config:                     options.di.MockConfigManager,
		Logger:                     options.di.NoopLogger,
		I18nBundle:                 options.di.I18nBundle,
		HTTPRoutes:                 options.routes,
		SessionManager:             options.di.SessionManager,
		Tracer:                     options.di.NoopTracer,
		UniversalTranslator:        options.di.UniversalTranslator,
		GetPreferredUserLanguageFn: options.di.GetPreferredUserLanguageFn,
	})
}
