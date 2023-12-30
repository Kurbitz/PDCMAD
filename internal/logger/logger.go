package logger

import (
	"io"
	"log/slog"
	"os"
)

func NewLogger() {
	//Open the logfile to write
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	//Log to both the file and the console
	logger := slog.New(slog.NewTextHandler((io.MultiWriter(file, os.Stdout)), nil))
	slog.SetDefault(logger)

}

// When outputting errors we don't want to write them twice to console
// This function ensures that does not happen
func OutputError(printErr error) {
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewTextHandler(file, nil))
	logger.Info(printErr.Error())

}
