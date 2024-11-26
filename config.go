package main

import (
	"log/slog"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/ninlil/envsubst"
	"gopkg.in/yaml.v2"
)

type cmdArgs struct {
	ConfigFile string `arg:"-c" help:"Config file to load" default:"config.yaml"`
}

var args cmdArgs

func loadConfig() error {
	err := arg.Parse(&args)
	if err != nil {
		return err
	}

	stat, err := os.Stat(args.ConfigFile)
	if err != nil {
		slog.Error("Failed to find config-file", "error", err)
		os.Exit(1)
	}
	if stat.IsDir() {
		slog.Error("Config-file is a directory", "error", err)
		os.Exit(1)
	}

	buf, err := os.ReadFile(args.ConfigFile)
	if err != nil {
		slog.Error("Failed to read config-file", "error", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		slog.Error("Failed to parse config-file", "error", err)
		os.Exit(1)
	}

	envsubst.SetPrefix('$')
	envsubst.SetWrapper('{')

	return nil
}
