package database

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLogHandler is a custom handler for capturing log output during tests
type testLogHandler struct {
	buf    *bytes.Buffer
	level  slog.Level
	logged []map[string]interface{}
}

func newTestLogHandler(buf *bytes.Buffer) *testLogHandler {
	return &testLogHandler{
		buf:    buf,
		level:  slog.LevelDebug,
		logged: make([]map[string]interface{}, 0),
	}
}

func (h *testLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *testLogHandler) Handle(_ context.Context, r slog.Record) error {
	logEntry := make(map[string]interface{})
	logEntry["level"] = r.Level.String()
	logEntry["message"] = r.Message
	
	r.Attrs(func(a slog.Attr) bool {
		logEntry[a.Key] = a.Value.String()
		return true
	})
	
	h.logged = append(h.logged, logEntry)
	
	// Also output as JSON for string inspection tests
	data, _ := json.Marshal(logEntry)
	h.buf.Write(data)
	h.buf.WriteByte('\n')
	
	return nil
}

func (h *testLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestNewPgxLogger(t *testing.T) {
	// Setup
	var buf bytes.Buffer
	handler := newTestLogHandler(&buf)
	logger := slog.New(handler)
	
	// Execute
	pgxLogger := NewPgxLogger(logger)
	
	// Verify
	assert.NotNil(t, pgxLogger)
	assert.NotNil(t, pgxLogger.logger)
	assert.NotNil(t, pgxLogger.newlinePattern)
	assert.NotNil(t, pgxLogger.levelMapping)
	
	// Check mapping of log levels
	assert.Equal(t, slog.LevelDebug, pgxLogger.levelMapping[tracelog.LogLevelTrace])
	assert.Equal(t, slog.LevelDebug, pgxLogger.levelMapping[tracelog.LogLevelDebug])
	assert.Equal(t, slog.LevelInfo, pgxLogger.levelMapping[tracelog.LogLevelInfo])
	assert.Equal(t, slog.LevelWarn, pgxLogger.levelMapping[tracelog.LogLevelWarn])
	assert.Equal(t, slog.LevelError, pgxLogger.levelMapping[tracelog.LogLevelError])
}

func TestPgxLoggerLog(t *testing.T) {
	// Setup
	var buf bytes.Buffer
	handler := newTestLogHandler(&buf)
	logger := slog.New(handler)
	pgxLogger := NewPgxLogger(logger)
	ctx := context.Background()
	
	// Test cases
	tests := []struct {
		name     string
		level    tracelog.LogLevel
		msg      string
		data     map[string]interface{}
		expected bool // whether log should be output
	}{
		{
			name:     "debug level",
			level:    tracelog.LogLevelDebug,
			msg:      "debug message",
			data:     map[string]interface{}{"time": time.Now().Format(time.RFC3339)},
			expected: true,
		},
		{
			name:     "info level",
			level:    tracelog.LogLevelInfo,
			msg:      "info message",
			data:     map[string]interface{}{"time": time.Now().Format(time.RFC3339)},
			expected: true,
		},
		{
			name:     "unsupported level",
			level:    tracelog.LogLevelNone,
			msg:      "should not be logged",
			data:     map[string]interface{}{"time": time.Now().Format(time.RFC3339)},
			expected: false,
		},
	}
	
	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logCountBefore := len(handler.logged)
			pgxLogger.Log(ctx, tc.level, tc.msg, tc.data)
			
			if tc.expected {
				assert.Equal(t, logCountBefore+1, len(handler.logged), "Log should have been recorded")
			} else {
				assert.Equal(t, logCountBefore, len(handler.logged), "Log should not have been recorded")
			}
		})
	}
}

func TestPgxLoggerCleanSQL(t *testing.T) {
	// Setup
	var buf bytes.Buffer
	logger := slog.New(newTestLogHandler(&buf))
	pgxLogger := NewPgxLogger(logger)
	
	// Test cases
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line SQL",
			input:    "SELECT * FROM users WHERE id = $1",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "multiline SQL",
			input:    "SELECT *\nFROM users\nWHERE id = $1",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "SQL with extra spaces",
			input:    "SELECT   *   FROM   users   WHERE   id   =   $1",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "SQL with Windows line endings",
			input:    "SELECT *\r\nFROM users\r\nWHERE id = $1",
			expected: "SELECT * FROM users WHERE id = $1",
		},
	}
	
	// Execute each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := pgxLogger.cleanSQL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}