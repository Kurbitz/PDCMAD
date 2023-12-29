package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func NewLogger() {
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Errorf("Failed to open logfile")
	}
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler((io.MultiWriter(file, os.Stdout)), opts))
	slog.SetDefault(logger)

}
func OutputError(printErr error) {
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Errorf("Failed to open logfile")
	}
	logger := slog.New(slog.NewTextHandler(file, nil))
	logger.Info(printErr.Error())

}
