package database

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
)

type gooseSlogger struct {
	logger *slog.Logger
}

var _ goose.Logger = (*gooseSlogger)(nil)

func (s gooseSlogger) Printf(format string, v ...any) {
	s.logger.Info(fmt.Sprintf(format, v...))
}

func (s gooseSlogger) Fatalf(format string, v ...any) {
	s.logger.Error(fmt.Sprintf(format, v...))
	//nolint:revive // The Fatalf function must terminate the program, so calling os.Exit here is necessary.
	os.Exit(1)
}
