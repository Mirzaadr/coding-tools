package builders

import (
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/filesystem"
	"github.com/andre/dotnet-architect/internal/templates"
)

// EngineTemplateRenderer is the production TemplateRenderer, delegating to
// templates.Engine for rendering and filesystem.FileSystem for writing.
type EngineTemplateRenderer struct {
	engine *templates.Engine
	fs     filesystem.FileSystem
	logger *slog.Logger
}

// NewEngineTemplateRenderer constructs an EngineTemplateRenderer.
func NewEngineTemplateRenderer(engine *templates.Engine, fs filesystem.FileSystem, logger *slog.Logger) *EngineTemplateRenderer {
	return &EngineTemplateRenderer{engine: engine, fs: fs, logger: logger}
}

func (r *EngineTemplateRenderer) RenderToFile(templateRelativePath string, data any, destinationPath string) error {
	r.logger.Debug("rendering template", "template", templateRelativePath, "destination", destinationPath)

	rendered, err := r.engine.Render(templateRelativePath, data)
	if err != nil {
		return fmt.Errorf("builders: failed to render template %q: %w", templateRelativePath, err)
	}

	if err := r.fs.WriteFile(destinationPath, rendered); err != nil {
		return fmt.Errorf("builders: failed to write rendered template to %q: %w", destinationPath, err)
	}

	return nil
}
