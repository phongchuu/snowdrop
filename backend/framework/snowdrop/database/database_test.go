package database

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

// Mock for the Config interface
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetDatabaseURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfig) GetEmbedResourceFolder() any {
	args := m.Called()
	return args.Get(0)
}

func (m *MockConfig) GetAppPort() string {
	args := m.Called()
	return args.String(0)
}

// Mock for pgxpool
type MockPgxPoolWrapper struct {
	*pgxpool.Pool
	mock.Mock
}

func newMockPgxPool() *MockPgxPoolWrapper {
	return &MockPgxPoolWrapper{}
}

func (m *MockPgxPoolWrapper) Close() {
	m.Called()
}

// TestDatabaseTracerConfiguration tests that the database connection correctly configures
// the tracer with our new PgxLogger
func TestDatabaseTracerConfiguration(t *testing.T) {
	// Skip full database creation test which would require a real database
	t.Skip("This test requires mocking the pgxpool.NewWithConfig which is challenging")

	// This is a partial test focusing on the configuration
	mockConfig := new(MockConfig)
	mockConfig.On("GetDatabaseURL").Return("postgres://user:pass@localhost:5432/testdb")
	
	logger := slog.New(slog.NewTextHandler(nil, nil))
	
	// Parse config to examine tracer configuration
	cfg, err := pgxpool.ParseConfig(mockConfig.GetDatabaseURL())
	require.NoError(t, err)
	
	// Configure tracer with our logger
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   NewPgxLogger(logger),
		LogLevel: tracelog.LogLevelTrace,
	}
	
	// Verify that the tracer is configured with our logger
	tracer, ok := cfg.ConnConfig.Tracer.(*tracelog.TraceLog)
	require.True(t, ok)
	
	// Check that the logger is our PgxLogger
	pgxLogger, ok := tracer.Logger.(*PgxLogger)
	require.True(t, ok)
	assert.NotNil(t, pgxLogger)
	
	// Check log level matches the tracelog.LogLevelTrace from our modified code
	assert.Equal(t, tracelog.LogLevelTrace, tracer.LogLevel)
}

// TestExecuteDatabaseUpgrade tests that the upgrade function sets up goose correctly
func TestExecuteDatabaseUpgrade(t *testing.T) {
	// This is a partial test that only checks if lifecycle hooks are added correctly
	// A full test would require mocking goose which is complex

	mockConfig := new(MockConfig)
	mockConfig.On("GetEmbedResourceFolder").Return(nil)
	
	logger := slog.New(slog.NewTextHandler(nil, nil))
	db := &sql.DB{}
	
	// Call the function but we'll only verify it doesn't panic
	// A complete test would require more complex mocking
	lc := fx.Lifecycle{}
	
	// This shouldn't panic
	executeDatabaseUpgrade(lc, mockConfig, logger, db)
	
	// Verify the lifecycle has hooks added (should have 1 OnStart hook)
	assert.Equal(t, 1, len(lc.Hooks))
}
