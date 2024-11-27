package main

import (
	"log/slog"
	"os"

	"github.com/AB-Lindex/http-rewrite/internal/slog2"
)

func init() {
	slog.SetDefault(slog.New(slog2.New(os.Stdout, &slog2.Options{Level: slog.LevelDebug})))
}
