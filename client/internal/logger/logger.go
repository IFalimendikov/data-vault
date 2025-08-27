package logger

import (
	"log/slog"
	"os"
)

// New creates a new structured logger with text output to stdout
func New() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return log
}
