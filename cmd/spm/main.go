// Package main
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/alexmatiasas/spm/internal/cmd"
	"github.com/alexmatiasas/spm/internal/config"
	"github.com/alexmatiasas/spm/internal/logging"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	levelName := config.LogLevel()
	if levelName == "" {
		levelName = slog.LevelInfo.String()
	}

	logging.Setup(os.Stderr, levelName)
	slog.Info("spm starting", "version", version, "commit", commit, "level", levelName)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.NewRootCmd(version, commit).ExecuteContext(ctx); err != nil {
		return 1
	}

	return 0
}
