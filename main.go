package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	if err := loadConfig(); err != nil {
		slog.Error("Failed loading config", "error", err)
		os.Exit(1)
	}

	r := chi.NewRouter()

	var nReg int
	for _, api := range cfg.APIs {
		err := api.Handle(r)
		if err != nil {
			slog.Error("Failed to register API", "path", api.Input.Path, "error", err)
		} else {
			nReg++
		}
	}
	if nReg == 0 {
		slog.Error("No APIs registered")
		os.Exit(1)
	}

	slog.Info("Starting server", "port", cfg.ListenPort)

	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.ListenPort), r)
	if err != nil {
		panic(err)
	}
}
