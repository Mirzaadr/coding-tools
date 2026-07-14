package builders

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/dotnet"
)

// DotnetReferenceBuilder is the production ReferenceBuilder, delegating to
// the dotnet.Dotnet wrapper.
type DotnetReferenceBuilder struct {
	dotnetClient dotnet.Dotnet
	logger       *slog.Logger
}

// NewDotnetReferenceBuilder constructs a DotnetReferenceBuilder.
func NewDotnetReferenceBuilder(dotnetClient dotnet.Dotnet, logger *slog.Logger) *DotnetReferenceBuilder {
	return &DotnetReferenceBuilder{dotnetClient: dotnetClient, logger: logger}
}

func (b *DotnetReferenceBuilder) AddProjectToSolution(ctx context.Context, solutionPath string, projectPath string) error {
	b.logger.Debug("adding project to solution", "solutionPath", solutionPath, "projectPath", projectPath)

	if _, err := b.dotnetClient.AddProjectToSolution(ctx, solutionPath, projectPath); err != nil {
		return fmt.Errorf("builders: failed to add project %q to solution %q: %w", projectPath, solutionPath, err)
	}

	return nil
}

func (b *DotnetReferenceBuilder) AddProjectReference(ctx context.Context, projectPath string, referencedProjectPath string) error {
	b.logger.Debug("adding project reference", "projectPath", projectPath, "referencedProjectPath", referencedProjectPath)

	if _, err := b.dotnetClient.AddReference(ctx, projectPath, referencedProjectPath); err != nil {
		return fmt.Errorf("builders: failed to add reference from %q to %q: %w", projectPath, referencedProjectPath, err)
	}

	return nil
}
