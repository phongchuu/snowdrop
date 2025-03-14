package config

import (
	"fmt"
	"io/fs"
	"net"

	"github.com/spf13/viper"
)

// AppConfig is a struct that holds the application's configuration.
type AppConfig struct {
	viperInstance        *viper.Viper
	embedResourcesFolder fs.FS
}

// GetEmbedResourceFolder returns the embedded resource folder as a filesystem (fs.FS).
// This folder contains the embedded resources used by the application.
func (appCfg AppConfig) GetEmbedResourceFolder() fs.FS {
	return appCfg.embedResourcesFolder
}

// GetDatabaseURL constructs and returns the database connection URL
// using the configuration values stored in the AppConfig instance.
// It retrieves the database username, password, host, port, and name
// from the viper instance and formats them into a PostgreSQL connection string.
//
// Returns: The formatted PostgreSQL connection URL.
func (appCfg AppConfig) GetDatabaseURL() string {
	username := appCfg.viperInstance.GetString("db_username")
	password := appCfg.viperInstance.GetString("db_password")
	dbHost := appCfg.viperInstance.GetString("db_host")
	dbName := appCfg.viperInstance.GetString("db_name")
	dbPort := appCfg.viperInstance.GetString("db_port")

	return fmt.Sprintf("postgresql://%s:%s@%s/%s", username, password, net.JoinHostPort(dbHost, dbPort), dbName)
}

// NewAppConfig initializes a new AppConfig instance with the provided embedded resources folder.
// It sets up a new Viper instance, configures it to use environment variables with the prefix "app",
// and returns the configured AppConfig instance.
//
// Parameters:
//   - embedResourcesFolder: an fs.FS instance representing the folder containing embedded resources.
//
// Returns:
//   - AppConfig: the initialized application configuration.
//   - error: an error if the configuration setup fails.
func NewAppConfig(embedResourcesFolder fs.FS) (AppConfig, error) {
	viperInstance := viper.New()
	viperInstance.SetEnvPrefix("app")
	viperInstance.AutomaticEnv()

	return AppConfig{
		viperInstance:        viperInstance,
		embedResourcesFolder: embedResourcesFolder,
	}, nil
}
