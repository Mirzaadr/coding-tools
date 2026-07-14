package builders

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/dotnet"
)

// DotnetPackageInstaller is the production PackageInstaller, delegating to
// the dotnet.Dotnet wrapper.
type DotnetPackageInstaller struct {
	dotnetClient dotnet.Dotnet
	logger       *slog.Logger
}

// NewDotnetPackageInstaller constructs a DotnetPackageInstaller.
func NewDotnetPackageInstaller(dotnetClient dotnet.Dotnet, logger *slog.Logger) *DotnetPackageInstaller {
	return &DotnetPackageInstaller{dotnetClient: dotnetClient, logger: logger}
}

func (b *DotnetPackageInstaller) InstallPackage(ctx context.Context, projectPath string, packageName string, version string) error {
	b.logger.Debug("installing package", "projectPath", projectPath, "package", packageName, "version", version)

	if _, err := b.dotnetClient.AddPackage(ctx, projectPath, packageName, version); err != nil {
		return fmt.Errorf("builders: failed to install package %q into %q: %w", packageName, projectPath, err)
	}

	return nil
}

func (b *DotnetPackageInstaller) InstallPackages(ctx context.Context, projectPath string, packages map[string]string) error {
	for name, version := range packages {
		if err := b.InstallPackage(ctx, projectPath, name, version); err != nil {
			return err
		}
	}

	return nil
}
