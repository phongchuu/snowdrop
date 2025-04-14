package database

import (
	"context"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

type PgxLogger struct {
	logger         *slog.Logger
	newlinePattern *regexp.Regexp
	levelMapping   map[tracelog.LogLevel]slog.Level
}

var _ tracelog.Logger = (*PgxLogger)(nil)

func NewPgxLogger(logger *slog.Logger) *PgxLogger {
	return &PgxLogger{
		logger:         logger,
		newlinePattern: regexp.MustCompile(`\r?\n`),
		//nolint:exhaustive // LogLevelNone is intentionally unsupported and omitted from this mapping.
		levelMapping: map[tracelog.LogLevel]slog.Level{
			tracelog.LogLevelTrace: slog.LevelDebug,
			tracelog.LogLevelDebug: slog.LevelDebug,
			tracelog.LogLevelInfo:  slog.LevelInfo,
			tracelog.LogLevelWarn:  slog.LevelWarn,
			tracelog.LogLevelError: slog.LevelError,
		},
	}
}

// Log implements tracelog.Logger.
func (p *PgxLogger) Log(
	ctx context.Context,
	level tracelog.LogLevel,
	msg string,
	data map[string]any,
) {
	if !p.isLevelSupported(level) {
		return
	}

	sql, attrs := p.processAttributes(data)
	logMessage := p.buildLogMessage(msg, data["time"], sql)

	p.logger.LogAttrs(ctx, p.levelMapping[level], logMessage, attrs...)
}

// isLevelSupported checks if the given log level is mapped in the logger.
func (p *PgxLogger) isLevelSupported(level tracelog.LogLevel) bool {
	_, exists := p.levelMapping[level]

	return exists
}

// processAttributes extracts SQL and converts data map to slog attributes.
func (p *PgxLogger) processAttributes(data map[string]any) (string, []slog.Attr) {
	var sql string

	attrs := lo.MapToSlice(data, func(key string, value any) slog.Attr {
		if key == "sql" {
			sql = p.cleanSQL(cast.ToString(value))

			return slog.Any(key, sql)
		}

		return slog.Any(key, value)
	})

	return sql, attrs
}

// cleanSQL normalizes SQL by removing newlines and extra whitespace.
func (p *PgxLogger) cleanSQL(rawSQL string) string {
	// Replace newlines with spaces
	noNewlines := p.newlinePattern.ReplaceAllString(rawSQL, " ")

	// Collapse multiple whitespace characters into single spaces
	return strings.Join(strings.Fields(noNewlines), " ")
}

// buildLogMessage constructs the formatted log message string.
func (p *PgxLogger) buildLogMessage(msg string, timeValue any, sql string) string {
	var builder strings.Builder

	builder.Grow(64) // Pre-allocate buffer for typical message size

	// Write timestamp and main message
	builder.WriteString("[")
	builder.WriteString(cast.ToString(timeValue))
	builder.WriteString("] [")
	builder.WriteString(msg)
	builder.WriteString("]")

	// Append SQL if present
	if sql != "" {
		builder.WriteString(" ")
		builder.WriteString(sql)
	}

	return builder.String()
}
