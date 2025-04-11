package log

import "log/slog"

// NewNoopLogger returns a logger that discards all log messages.
// It can be used for testing or when no logging is required.
func NewNoopLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
