package log

import (
	"log/slog"
	"os"
	"path/filepath"
)

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
