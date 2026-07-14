// Package cli wires Cobra commands to the underlying business logic
// packages (models, architecture, services, ...). It is the only package
// allowed to import Cobra; everything it depends on is plain Go designed to
// be usable without a CLI at all.
package cli

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/andre/dotnet-architect/internal/cli/commands"
	"github.com/andre/dotnet-architect/internal/cli/config"
	"github.com/andre/dotnet-architect/internal/utils"
)

// NewRootCommand builds the `arch` root command with all subcommands
// registered. Adding a new subcommand means adding one line here - nothing
// else in this function needs to change.
func NewRootCommand() (*cobra.Command, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("cli: failed to load config: %w", err)
	}

	var verbose bool

	rootCmd := &cobra.Command{
		Use:   "arch",
		Short: "DotnetArchitect generates opinionated .NET project templates",
		Long: "DotnetArchitect is a CLI tool that generates opinionated .NET project\n" +
			"templates based on different software architectures (Clean Architecture,\n" +
			"and more to come).",
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose (debug) logging")

	// logger is constructed lazily via PersistentPreRun so the --verbose
	// flag (parsed by Cobra) is honored.
	var logger *slog.Logger
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		logger = utils.NewLogger(verbose)
	}

	rootCmd.AddCommand(commands.NewCreateCommand(func() *slog.Logger {
		if logger == nil {
			logger = utils.NewLogger(verbose)
		}
		return logger
	}, cfg))

	return rootCmd, nil
}
