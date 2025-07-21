package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type (
	LogsConfig struct {
		Level  string `koanf:"level"`
		Source bool   `koanf:"source"`
		Format string `koanf:"format"`
	}
	Logger struct {
		slogLogger *slog.Logger
	}
)

var (
	log               *slog.Logger
	level             *slog.LevelVar
	defaultLogsConfig = LogsConfig{
		Level:  "debug",
		Source: true,
		Format: "text",
	}
)

func GetLogger(l *slog.Logger) *Logger {
	return &Logger{slogLogger: l}
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.slogLogger.Debug("fatal", "error", fmt.Sprintf(format, v...))
	os.Exit(1)
}
func (l *Logger) Printf(format string, v ...any) {
	l.slogLogger.Debug("", "msg", fmt.Sprintf(format, v...))
}

func parseLevel(level string) (slog.Level, error) {
	l := strings.ToLower(level)
	switch l {
	case "err", "error":
		return slog.LevelError, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	}
	return slog.LevelDebug, errors.New("can't parse log level")
}

func SetupLogs(config LogsConfig) {
	var l slog.Level
	var h slog.Handler
	var err error

	if l, err = parseLevel(config.Level); err != nil {
		//nolint:forbidigo // Allow printing usage
		fmt.Printf("Can't read log level, defaulting to debug\n")
	}
	level = new(slog.LevelVar)
	// if config.Format == "text" {
	options := tint.Options{
		Level:      level,
		TimeFormat: "15:04:05.000",
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
	}
	h = tint.NewHandler(os.Stdout, &options)
	// }
	log = slog.New(h)
	level.Set(l)
}

func UpdateLogs(config LogsConfig) {
	var l slog.Level
	var err error

	if l, err = parseLevel(config.Level); err != nil {
		log.Error("Can't parse log level, not changing it.", "newlevel", config.Level)
		return
	}
	level.Set(l)
}
