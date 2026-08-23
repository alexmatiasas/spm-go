// Package main
package main

import (
	"fmt"
	"log/slog"
	"os"

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

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("spm %s (commit %s)\n", version, commit)
		return
	}

	slog.Debug("no subcommand given, showing usage")
	fmt.Println("spm — Smart Project Manager")
	fmt.Println("Run 'spm --help' for usage information.")
}
