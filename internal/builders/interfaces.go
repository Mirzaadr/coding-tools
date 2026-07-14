// Package builders contains small, single-purpose collaborators that
// ArchitectureGenerator implementations orchestrate. No builder depends on
// Cobra or any CLI concern; each depends only on the dotnet, filesystem,
// and templates packages via constructor injection.
package builders

import "context"

// SolutionBuilder creates the top-level .sln file for a generated project.
type SolutionBuilder interface {
	Build(ctx context.Context, name string, outputPath string) error
}

// ProjectBuilder creates an individual .csproj project (class library,
// web api, mvc app, etc.) within a solution.
type ProjectBuilder interface {
	BuildClassLibrary(ctx context.Context, name string, outputPath string) (string, error)
	BuildWebApi(ctx context.Context, name string, outputPath string) (string, error)
	BuildMvc(ctx context.Context, name string, outputPath string) (string, error)
}

// ReferenceBuilder wires up solution membership and project-to-project
// references.
type ReferenceBuilder interface {
	AddProjectToSolution(ctx context.Context, solutionPath string, projectPath string) error
	AddProjectReference(ctx context.Context, projectPath string, referencedProjectPath string) error
}

// FolderBuilder creates the conventional folder layout inside a generated
// project (e.g. Domain/Entities, Application/Interfaces, etc.).
type FolderBuilder interface {
	CreateFolders(ctx context.Context, basePath string, relativeFolders []string) error
}

// PackageInstaller adds NuGet packages to a project, e.g. EF Core provider
// packages selected based on the chosen DatabaseType.
type PackageInstaller interface {
	InstallPackage(ctx context.Context, projectPath string, packageName string, version string) error
	InstallPackages(ctx context.Context, projectPath string, packages map[string]string) error
}

// TemplateRenderer renders a .tmpl file to a data context and writes the
// result to disk.
type TemplateRenderer interface {
	RenderToFile(templateRelativePath string, data any, destinationPath string) error
}
