// Command arch is the entrypoint for DotnetArchitect.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/andre/dotnet-architect/internal/cli"
)

func main() {
	os.Args = normalizeDBShorthand(os.Args)

	rootCmd, err := cli.NewRootCommand()
	if err != nil {
		fmt.Fprintln(os.Stderr, "arch: fatal:", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// normalizeDBShorthand rewrites the documented "-db" short flag into the
// "--database" long flag pflag actually understands, since pflag (unlike
// dotnet's own CLI parser) only supports single-character shorthands.
// Handles both "-db PostgreSQL" and "-db=PostgreSQL" forms.
func normalizeDBShorthand(args []string) []string {
	normalized := make([]string, 0, len(args))
	for _, arg := range args {
		switch {
		case arg == "-db":
			normalized = append(normalized, "--database")
		case strings.HasPrefix(arg, "-db="):
			normalized = append(normalized, "--database="+strings.TrimPrefix(arg, "-db="))
		default:
			normalized = append(normalized, arg)
		}
	}
	return normalized
}
