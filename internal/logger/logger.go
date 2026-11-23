package logger

import (
	"context"
	"log/slog"
	"os"
)

func Setup(level string) error {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})))
	return nil
}

type Frontend struct{}

func (Frontend) Write(level, message string) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	slog.Log(context.Background(), lvl, message, "source", "frontend")
}
