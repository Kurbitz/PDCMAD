package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	slog.Logger
}

func NewLogger() *slog.Logger {
	file, err := os.Create("log.txt")
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewTextHandler(file, nil))
	return logger
}

func (log Logger) Info() {

}

func (log Logger) Warn() {

}
