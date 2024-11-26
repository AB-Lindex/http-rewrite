package main

import (
	"log/slog"
	"os"

	"github.com/AB-Lindex/http-rewrite/internal/slog2"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.SetDefault(slog.New(slog2.New(os.Stdout, nil)))
}
