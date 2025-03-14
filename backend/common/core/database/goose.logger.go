package database

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
)

type gooseSlogger struct {
	Logger *slog.Logger
}

var _ goose.Logger = (*gooseSlogger)(nil)

func (s gooseSlogger) Printf(format string, v ...any) {
	s.Logger.Info(fmt.Sprintf(format, v...))
}

func (s gooseSlogger) Fatalf(format string, v ...any) {
	s.Logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}
