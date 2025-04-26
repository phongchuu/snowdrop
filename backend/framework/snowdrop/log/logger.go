package log

import (
	"log/slog"
)

// ErrorLogAttr creates an slog attribute with the key "details" and the provided error value.
// This function is useful for adding error details to log records.
func ErrorLogAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func StacktraceLogAttr(stacktrace string) slog.Attr {
	return slog.String("stacktrace", stacktrace)
}
