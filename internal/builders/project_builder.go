package builders

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/dotnet"
	"github.com/andre/dotnet-architect/internal/filesystem"
)

// DotnetProjectBuilder is the production ProjectBuilder, delegating to the
// dotnet.Dotnet wrapper.
type DotnetProjectBuilder struct {
	dotnetClient dotnet.Dotnet
	fs           filesystem.FileSystem
	logger       *slog.Logger
}

// NewDotnetProjectBuilder constructs a DotnetProjectBuilder.
func NewDotnetProjectBuilder(dotnetClient dotnet.Dotnet, fs filesystem.FileSystem, logger *slog.Logger) *DotnetProjectBuilder {
	return &DotnetProjectBuilder{dotnetClient: dotnetClient, fs: fs, logger: logger}
}

func (b *DotnetProjectBuilder) BuildClassLibrary(ctx context.Context, name string, outputPath string) (string, error) {
	b.logger.Debug("building class library project", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateClassLibrary(ctx, name, outputPath); err != nil {
		return "", fmt.Errorf("builders: failed to create class library %q: %w", name, err)
	}

	return b.fs.Join(outputPath, name+".csproj"), nil
}

func (b *DotnetProjectBuilder) BuildWebApi(ctx context.Context, name string, outputPath string) (string, error) {
	b.logger.Debug("building web api project", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateWebApi(ctx, name, outputPath); err != nil {
		return "", fmt.Errorf("builders: failed to create web api %q: %w", name, err)
	}

	return b.fs.Join(outputPath, name+".csproj"), nil
}

func (b *DotnetProjectBuilder) BuildMvc(ctx context.Context, name string, outputPath string) (string, error) {
	b.logger.Debug("building mvc project", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateMvc(ctx, name, outputPath); err != nil {
		return "", fmt.Errorf("builders: failed to create mvc app %q: %w", name, err)
	}

	return b.fs.Join(outputPath, name+".csproj"), nil
}
