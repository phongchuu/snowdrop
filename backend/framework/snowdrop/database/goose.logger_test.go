package database

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExit replaces os.Exit during tests
type mockExit struct {
	code int
	called bool
}

func (m *mockExit) exit(code int) {
	m.code = code
	m.called = true
}

// captureOutput captures output from slog to a buffer
func captureOutput(t *testing.T) (*bytes.Buffer, *slog.Logger) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	return &buf, logger
}

func TestGooseSloggerImplementsInterface(t *testing.T) {
	// Setup
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	
	// Create the goose logger
	glogger := &gooseSlogger{logger: logger}
	
	// Verify it implements the goose.Logger interface
	// This is mostly a compile-time check, but we'll assert it's not nil
	assert.NotNil(t, glogger)
	
	// Check that the logger field is correctly set
	assert.Equal(t, logger, glogger.logger)
}

func TestGooseSloggerPrintf(t *testing.T) {
	// Setup
	buf, logger := captureOutput(t)
	glogger := &gooseSlogger{logger: logger}
	
	// Execute
	testMessage := "test message %d"
	glogger.Printf(testMessage, 42)
	
	// Verify the output contains our message and is logged at info level
	output := buf.String()
	assert.Contains(t, output, "test message 42")
	assert.Contains(t, output, `"level":"INFO"`)
}

func TestGooseSloggerFatalf(t *testing.T) {
	if os.Getenv("TEST_FATALF") == "1" {
		// When running in subprocess mode, execute the actual code
		var buf bytes.Buffer
		handler := slog.NewJSONHandler(&buf, nil)
		logger := slog.New(handler)
		glogger := &gooseSlogger{logger: logger}
		glogger.Fatalf("fatal message %d", 42)
		return
	}
	
	// Create a subprocess to test the os.Exit behavior
	cmd := exec.Command(os.Args[0], "-test.run=TestGooseSloggerFatalf")
	cmd.Env = append(os.Environ(), "TEST_FATALF=1")
	
	// Capture output
	output, err := cmd.CombinedOutput()
	
	// The process should exit with code 1
	e, ok := err.(*exec.ExitError)
	if assert.True(t, ok, "Expected process to exit with error") {
		assert.Equal(t, 1, e.ExitCode(), "Expected exit code 1")
	}
	
	// Check that the output contains our message and is at error level
	outputStr := string(output)
	assert.Contains(t, outputStr, "fatal message 42")
	assert.Contains(t, outputStr, `"level":"ERROR"`)
}
