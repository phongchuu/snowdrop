package snowdrop

import (
	"io/fs"

	"golang.org/x/text/language"
)

// ConfigManager provides an interface for managing application configuration settings.
type ConfigManager interface {
	// GetAppVersion returns the application version
	GetAppVersion() string

	// GetAppRevision returns the application revision/build number
	GetAppRevision() string

	// GetEmbedResourceFolder returns the embedded resource folder as a filesystem (fs.FS).
	// This folder contains the embedded resources used by the application.
	GetEmbedResourceFolder() fs.FS

	// GetDatabaseURL constructs and returns the database connection URL
	// using the configuration values stored in the AppConfig instance.
	// It retrieves the database username, password, host, port, and name
	// from the viper instance and formats them into a PostgreSQL connection string.
	GetDatabaseURL() string

	// GetAppPort returns the port number the application should listen on
	GetAppPort() string

	// GetSupportedLanguages returns a list of supported language tags.
	GetSupportedLanguages() []language.Tag

	// GetMaxRequestSize returns the maximum request size in bytes.
	GetMaxRequestSize() int64
}
