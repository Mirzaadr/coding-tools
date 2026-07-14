package builders

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/filesystem"
)

// StandardFolderBuilder is the production FolderBuilder, delegating to the
// filesystem.FileSystem abstraction.
type StandardFolderBuilder struct {
	fs     filesystem.FileSystem
	logger *slog.Logger
}

// NewStandardFolderBuilder constructs a StandardFolderBuilder.
func NewStandardFolderBuilder(fs filesystem.FileSystem, logger *slog.Logger) *StandardFolderBuilder {
	return &StandardFolderBuilder{fs: fs, logger: logger}
}

func (b *StandardFolderBuilder) CreateFolders(ctx context.Context, basePath string, relativeFolders []string) error {
	for _, relative := range relativeFolders {
		fullPath := b.fs.Join(basePath, relative)

		b.logger.Debug("creating folder", "path", fullPath)

		if err := b.fs.MkdirAll(fullPath); err != nil {
			return fmt.Errorf("builders: failed to create folder %q: %w", fullPath, err)
		}
	}

	return nil
}
