package testutils

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"

	"github.com/go-chi/chi/v5"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	SessionManager             snowdrop.SessionManager
	MockUserService            *mockusermgt.MockUserService
	GetPreferredUserLanguageFn snowdrop.GetPreferredUserLanguageFn
	DB                         *sql.DB
	PgContainer                *postgres.PostgresContainer
}

// StartPostgresContainer starts a PostgreSQL container for tests.
// The container is reusable and a snapshot is created after startup.
func StartPostgresContainer(ctx context.Context) *postgres.PostgresContainer {
	GinkgoHelper()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:17.4-alpine3.21",
		postgres.WithDatabase("snowdrop"),
		postgres.WithUsername("tester"),
		postgres.WithPassword("Keep!t5ecret"),
		postgres.BasicWaitStrategies(),
		testcontainers.WithReuseByName("e2e-postgresql"),
	)
	Expect(err).NotTo(HaveOccurred(), "Failed to start test postgres container")

	//nolint:contextcheck // False positive
	DeferCleanup(func(cleanupContextSpec SpecContext) error {
		return RestorePgContainer(cleanupContextSpec, pgContainer)
	})

	err = pgContainer.Snapshot(ctx)
	Expect(err).NotTo(HaveOccurred(), "Failed to create snapshot of test postgres container")

	return pgContainer
}

func RestorePgContainer(ctx context.Context, pgContainer *postgres.PostgresContainer) error {
	if pgContainer.IsRunning() {
		err := pgContainer.Restore(ctx)

		return err
	}

	return nil
}

func IntegrationTestSetup(tb testing.TB) TestContainer {
	GinkgoHelper()

	// This is used for database migration
	tb.Setenv("APP_DEFAULT_SYSTEM_PASSWORD", "Keep!t5ecret")

	pgContainer := StartPostgresContainer(tb.Context())

	fxLifecycle := fxtest.NewLifecycle(tb)
	noopLogger := log.NewNoopLogger()
	noopTracer := noop.NewTracerProvider().Tracer("noop.tracer")

	// ConfigManager
	mockConfigManager := mocksnowdrop.NewMockConfigManager(tb)
	mockConfigManager.EXPECT().GetEmbedResourceFolder().Return(GetResourceFS())
	mockConfigManager.EXPECT().GetSupportedLanguages().Return([]language.Tag{
		language.English,
		language.Vietnamese,
	})
	mockConfigManager.EXPECT().GetMaxRequestSize().Return(int64(1 << 20)).Maybe()
	mockConfigManager.EXPECT().
		GetDatabaseURL().
		Return(pgContainer.MustConnectionString(tb.Context()))

	// I18nBundle
	i18nBundle, err := translation.NewI18nBundle(mockConfigManager)
	Expect(err).NotTo(HaveOccurred())

	// Validator
	validationModule, err := validation.NewValidator()
	Expect(err).NotTo(HaveOccurred())

	// GetPreferredUserLanguageFn
	getPreferredUserLanguageFn := core.NewGetPreferredUserLanguageFn(mockConfigManager)

	// Database
	sqlDB, err := database.NewDatabase(fxLifecycle, noopLogger, mockConfigManager)
	Expect(err).NotTo(HaveOccurred())

	database.ExecuteDatabaseUpgrade(fxLifecycle, noopLogger, sqlDB, mockConfigManager)

	// Start the lifecycle
	fxLifecycle.RequireStart()
	DeferCleanup(func() {
		fxLifecycle.RequireStop()
	})

	// TransactionManager
	transactionManager := database.NewTransactionManager(sqlDB)

	// SessionManager
	sessionManager := session.NewManager(session.ManagerParams{
		TransactionManager: transactionManager,
		Tracer:             noopTracer,
		Repository:         session.NewPostgresRepository(transactionManager, noopTracer),
	})

	return TestContainer{
		GetPreferredUserLanguageFn: getPreferredUserLanguageFn,
		NoopLogger:                 noopLogger,
		MockConfigManager:          mockConfigManager,
		I18nBundle:                 i18nBundle,
		Validator:                  validationModule.Validator,
		UniversalTranslator:        validationModule.UniversalTranslator,
		NoopTracer:                 noopTracer,
		TransactionManager:         transactionManager,
		SessionManager:             sessionManager,
		DB:                         sqlDB,
		PgContainer:                pgContainer,
	}
}

func NewRouter(testContainer TestContainer, routes []snowdrop.HTTPHandler) *chi.Mux {
	GinkgoHelper()

	return web.NewRouter(web.RouteParams{
		Config:                     testContainer.MockConfigManager,
		Logger:                     testContainer.NoopLogger,
		I18nBundle:                 testContainer.I18nBundle,
		HTTPRoutes:                 routes,
		SessionManager:             testContainer.SessionManager,
		Tracer:                     testContainer.NoopTracer,
		UniversalTranslator:        testContainer.UniversalTranslator,
		GetPreferredUserLanguageFn: testContainer.GetPreferredUserLanguageFn,
	})
}
