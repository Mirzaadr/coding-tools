package builders

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

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

// defaultScaffoldFiles maps a dotnet template name to the file(s) it
// generates that we always want removed, since they get replaced by our
// own templates or aren't wanted in a generated architecture skeleton.
var defaultScaffoldFiles = map[string][]string{
	"classlib": {"Class1.cs"},
	"webapi":   {"WeatherForecast.cs", filepath.Join("Controllers", "WeatherForecastController.cs")},
	// "mvc" intentionally omitted for now - its scaffolding (HomeController,
	// Views/, wwwroot/) is more often kept than a class library's Class1.cs.
	// Add an entry here if you want it stripped too.
}

func (b *DotnetProjectBuilder) removeScaffoldFiles(template string, outputPath string) {
	for _, relative := range defaultScaffoldFiles[template] {
		full := b.fs.Join(outputPath, relative)
		if err := b.fs.Remove(full); err != nil {
			// Non-fatal: SDK output layout can vary by version. Log and move on
			// rather than failing generation over a cosmetic cleanup step.
			b.logger.Warn("failed to remove default scaffold file", "path", full, "error", err)
		}
	}
}

func (b *DotnetProjectBuilder) BuildClassLibrary(ctx context.Context, name string, outputPath string) (string, error) {
	b.logger.Debug("building class library project", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateClassLibrary(ctx, name, outputPath); err != nil {
		return "", fmt.Errorf("builders: failed to create class library %q: %w", name, err)
	}

	b.removeScaffoldFiles("classlib", outputPath)
	return b.fs.Join(outputPath, name+".csproj"), nil
}

func (b *DotnetProjectBuilder) BuildWebApi(ctx context.Context, name string, outputPath string) (string, error) {
	b.logger.Debug("building web api project", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateWebApi(ctx, name, outputPath); err != nil {
		return "", fmt.Errorf("builders: failed to create web api %q: %w", name, err)
	}

	b.removeScaffoldFiles("webapi", outputPath)
	return b.fs.Join(outputPath, name+".csproj"), nil
}

func (b *DotnetProjectBuilder) BuildMvc(ctx context.Context, name string, outputPath string) (string, error) {
	b.logger.Debug("building mvc project", "name", name, "outputPath", outputPath)

	if _, err := b.dotnetClient.CreateMvc(ctx, name, outputPath); err != nil {
		return "", fmt.Errorf("builders: failed to create mvc app %q: %w", name, err)
	}

	return b.fs.Join(outputPath, name+".csproj"), nil
}
