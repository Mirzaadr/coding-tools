// Package utils contains small, dependency-free helpers shared across the
// application. It must never import internal/cli, to keep business logic
// independent of the CLI framework.
package utils

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger builds the application's structured logger. All non-CLI-output
// logging should flow through a *slog.Logger obtained from here (or passed
// down via constructor injection) rather than fmt.Println.
func NewLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}

// NewSilentLogger returns a logger that discards all output. Useful for
// tests that don't care about log side effects.
func NewSilentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
