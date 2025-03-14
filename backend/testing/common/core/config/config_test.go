package core_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"internal.snowdrop/common/core/config"
)

func TestNewConfig(t *testing.T) {
	appCfg, err := config.NewAppConfig(fstest.MapFS{})
	require.NoError(t, err)
	assert.NotNil(t, appCfg)
}

func TestGetDatabaseURL(t *testing.T) {
	var embedFS fstest.MapFS

	t.Setenv("APP_DB_USERNAME", "testuser")
	t.Setenv("APP_DB_PASSWORD", "testpass")
	t.Setenv("APP_DB_HOST", "localhost")
	t.Setenv("APP_DB_NAME", "testdb")
	t.Setenv("APP_DB_PORT", "5432")

	appCfg, err := config.NewAppConfig(embedFS)
	require.NoError(t, err)
	assert.NotNil(t, appCfg)

	assert.Equal(t, "postgresql://testuser:testpass@localhost:5432/testdb", appCfg.GetDatabaseURL())
}

func TestGetEmbedResourceFolder(t *testing.T) {
	var embedFS fstest.MapFS

	appCfg, err := config.NewAppConfig(embedFS)
	require.NoError(t, err)
	assert.NotNil(t, appCfg)

	assert.Equal(t, embedFS, appCfg.GetEmbedResourceFolder())
}
