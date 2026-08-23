// Package main
package main

import (
	"log/slog"
	"os"

	"github.com/alexmatiasas/spm/internal/cmd"
	"github.com/alexmatiasas/spm/internal/config"
	"github.com/alexmatiasas/spm/internal/logging"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	levelName := config.LogLevel()
	if levelName == "" {
		levelName = slog.LevelInfo.String()
	}

	logging.Setup(os.Stderr, levelName)
	slog.Info("spm starting", "version", version, "commit", commit, "level", levelName)

	if err := cmd.NewRootCmd(version, commit).Execute(); err != nil {
		os.Exit(1)
	}
}
