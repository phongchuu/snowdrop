package core

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type slogFieldsCtxKey int

type contextHandler struct {
	slog.Handler
}

const slogFieldsCtxID slogFieldsCtxKey = 0

// Handle adds contextual attributes to the Record before calling the underlying.
func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogFieldsCtxID).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}

	return h.Handler.Handle(ctx, r)
}

// WithLogAttr adds an slog attribute to the provided context so that it will be
// included in any Record created with such context.
func WithLogAttr(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFieldsCtxID).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(parent, slogFieldsCtxID, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)

	return context.WithValue(parent, slogFieldsCtxID, v)
}

func WithLogLevel(level slog.Level) func(*slog.HandlerOptions) {
	return func(handlerOptions *slog.HandlerOptions) {
		handlerOptions.Level = level
	}
}

// newLogger creates a new slog logger with the specified writer and handler options.
// It uses the slog.NewJSONHandler to format log messages as JSON.
func newLogger(writer io.Writer, cfg *slog.HandlerOptions) *slog.Logger {
	return slog.New(&contextHandler{slog.NewJSONHandler(writer, cfg)})
}

// NewNoopLogger returns a logger that discards all log messages.
// It can be used for testing or when no logging is required.
func NewNoopLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// NewStdoutLogger returns a logger that writes log messages to the standard output (os.Stdout).
func NewStdoutLogger(opts ...func(*slog.HandlerOptions)) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, _ := a.Value.Any().(*slog.Source)
				a.Value = slog.AnyValue(slog.Source{
					Function: source.Function,
					File:     filepath.Base(source.File),
					Line:     source.Line,
				})
			}

			return a
		},
	}

	for _, optFn := range opts {
		optFn(handlerOptions)
	}

	return newLogger(os.Stdout, handlerOptions)
}

// ErrorLogAttr creates an slog attribute with the key "details" and the provided error value.
// This function is useful for adding error details to log records.
func ErrorLogAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
