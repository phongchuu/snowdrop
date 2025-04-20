package core

import (
	"fmt"
	"io/fs"
	"net"

	"github.com/spf13/viper"
	"golang.org/x/text/language"
	snowdrop "internal.snowdrop/framework"
)

// AppConfig is a struct that holds the application's configuration.
type AppConfig struct {
	// Common
	version            string
	revision           string
	supportedLanguages []language.Tag

	// Database
	databaseURL string

	// Embed folder
	embedResourcesFolder fs.FS

	// HTTP server
	httpPort string
}

var _ snowdrop.ConfigManager = (*AppConfig)(nil)

func (appCfg AppConfig) GetEmbedResourceFolder() fs.FS {
	return appCfg.embedResourcesFolder
}

func (appCfg AppConfig) GetDatabaseURL() string {
	return appCfg.databaseURL
}

func (appCfg AppConfig) GetAppPort() string {
	return appCfg.httpPort
}

func (appCfg AppConfig) GetAppVersion() string {
	return appCfg.version
}

func (appCfg AppConfig) GetAppRevision() string {
	return appCfg.revision
}

func (appCfg AppConfig) GetSupportedLanguages() []language.Tag {
	return appCfg.supportedLanguages
}

func (appCfg AppConfig) GetMaxRequestSize() int64 {
	return 1 << 20 // 1 MB
}

func initHTTPServerConfig(provider *viper.Viper, appCfg *AppConfig) {
	port := provider.GetString("app_port")

	if len(port) == 0 {
		port = "3000"
	}

	appCfg.httpPort = port
}

func initDatabaseConfig(provider *viper.Viper, appCfg *AppConfig) {
	username := provider.GetString("db_username")
	password := provider.GetString("db_password")
	dbHost := provider.GetString("db_host")
	dbName := provider.GetString("db_name")
	dbPort := provider.GetString("db_port")

	appCfg.databaseURL = fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		username,
		password,
		net.JoinHostPort(dbHost, dbPort),
		dbName,
	)
}

// NewAppConfig initializes and returns a new application configuration.
// It sets up environment variables with "app" prefix and default configuration values.
func NewAppConfig(moduleConfig ModuleConfig) (AppConfig, error) {
	// Create a new Viper instance
	viperInstance := viper.New()
	viperInstance.SetEnvPrefix("app")
	viperInstance.AutomaticEnv()

	config := AppConfig{
		version:  moduleConfig.Version,
		revision: moduleConfig.Revision,
		supportedLanguages: []language.Tag{
			language.English,    // en fallback language
			language.Vietnamese, // vi
		},
		embedResourcesFolder: moduleConfig.EmbedResourcesFolder,
	}

	initHTTPServerConfig(viperInstance, &config)
	initDatabaseConfig(viperInstance, &config)

	return config, nil
}
