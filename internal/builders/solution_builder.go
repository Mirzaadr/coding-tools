package builders

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/dotnet"
)

// DotnetSolutionBuilder is the production SolutionBuilder, delegating to
// the dotnet.Dotnet wrapper.
type DotnetSolutionBuilder struct {
	dotnetClient dotnet.Dotnet
	logger       *slog.Logger
}

// NewDotnetSolutionBuilder constructs a DotnetSolutionBuilder.
func NewDotnetSolutionBuilder(dotnetClient dotnet.Dotnet, logger *slog.Logger) *DotnetSolutionBuilder {
	return &DotnetSolutionBuilder{dotnetClient: dotnetClient, logger: logger}
}

func (b *DotnetSolutionBuilder) Build(ctx context.Context, name string, outputPath string) error {
	b.logger.Debug("building solution", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateSolution(ctx, name, outputPath); err != nil {
		return fmt.Errorf("builders: failed to create solution %q: %w", name, err)
	}

	return nil
}
