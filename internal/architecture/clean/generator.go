// Package clean implements the Clean Architecture ArchitectureGenerator.
// It contains no dotnet/os/exec calls of its own - it only orchestrates
// the shared builders package.
package clean

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andre/dotnet-architect/internal/architecture"
	"github.com/andre/dotnet-architect/internal/builders"
	"github.com/andre/dotnet-architect/internal/dotnet"
	"github.com/andre/dotnet-architect/internal/filesystem"
	"github.com/andre/dotnet-architect/internal/models"
)

// project layer names, used both as project name suffixes and as solution
// folders.
const (
	layerDomain         = "Domain"
	layerApplication    = "Application"
	layerInfrastructure = "Infrastructure"
)

// commonFolders are created inside each layer project as part of the
// conventional Clean Architecture folder layout.
var commonFolders = map[string][]string{
	layerDomain:         {"Entities", "Enums", "Exceptions", "Interfaces", "ValueObjects"},
	layerApplication:    {"Common"},
	layerInfrastructure: {"Persistence", "Services"},
}

// efCorePackages maps a DatabaseType to the NuGet package(s) required for
// EF Core to talk to it.
var efCorePackages = map[models.DatabaseType]map[string]string{
	models.DatabasePostgreSQL: {"Npgsql.EntityFrameworkCore.PostgreSQL": ""},
	models.DatabaseSqlServer:  {"Microsoft.EntityFrameworkCore.SqlServer": ""},
}

// Generator implements architecture.ArchitectureGenerator for Clean
// Architecture. All dependencies are injected explicitly via the
// constructor - no globals, no service locator.
type Generator struct {
	solutionBuilder  builders.SolutionBuilder
	projectBuilder   builders.ProjectBuilder
	referenceBuilder builders.ReferenceBuilder
	folderBuilder    builders.FolderBuilder
	packageInstaller builders.PackageInstaller
	templateRenderer builders.TemplateRenderer
	fs               filesystem.FileSystem
	dotnetClient     dotnet.Dotnet
	logger           *slog.Logger
	onProgress       architecture.ProgressFunc
}

// NewGenerator constructs a Clean Architecture Generator from its
// collaborators. onProgress may be nil, in which case progress reporting
// is a no-op.
func NewGenerator(
	solutionBuilder builders.SolutionBuilder,
	projectBuilder builders.ProjectBuilder,
	referenceBuilder builders.ReferenceBuilder,
	folderBuilder builders.FolderBuilder,
	packageInstaller builders.PackageInstaller,
	templateRenderer builders.TemplateRenderer,
	fs filesystem.FileSystem,
	dotnetClient dotnet.Dotnet,
	logger *slog.Logger,
	onProgress architecture.ProgressFunc,
) *Generator {
	if onProgress == nil {
		onProgress = func(string) {} // no-op default
	}
	return &Generator{
		solutionBuilder:  solutionBuilder,
		projectBuilder:   projectBuilder,
		referenceBuilder: referenceBuilder,
		folderBuilder:    folderBuilder,
		packageInstaller: packageInstaller,
		templateRenderer: templateRenderer,
		fs:               fs,
		dotnetClient:     dotnetClient,
		logger:           logger,
		onProgress:       onProgress,
	}
}

// Name identifies this generator in the architecture.Registry.
func (g *Generator) Name() models.ArchitectureType {
	return models.ArchitectureClean
}

// Generate produces a Clean Architecture solution on disk:
//
//	Solution
//	├── Domain
//	├── Application
//	├── Infrastructure
//	└── Presentation (WebApi or WebApp)
//
// with project references wired Presentation -> Application -> Domain and
// Infrastructure -> Application/Domain, EF Core packages installed based on
// options.Database, and a final restore/build unless options.Build is false.
//
// NOTE: this is the full intended behavior for a future phase. It is not
// yet invoked by the `arch create` command, which today only resolves and
// prints the configuration (see internal/services and the create command).
func (g *Generator) Generate(ctx context.Context, options models.ProjectOptions) error {
	root := options.OutputPath
	if options.CreateProjectFolder {
		root = g.fs.Join(options.OutputPath, options.Name)
	}

	g.onProgress("Preparing output directory")
	if err := g.folderBuilder.CreateFolders(ctx, root, []string{}); err != nil {
		return fmt.Errorf("clean: failed to prepare root directory: %w", err)
	}

	g.onProgress("Creating solution")
	if err := g.solutionBuilder.Build(ctx, options.Name, root); err != nil {
		return fmt.Errorf("clean: failed to build solution: %w", err)
	}

	solutionPath := g.fs.Join(root, options.Name+".sln")
	if !g.fs.Exists(solutionPath) {
		solutionPath = g.fs.Join(root, options.Name+".slnx")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("clean: cancelled: %w", err)
	}

	g.onProgress("Building Domain project")
	domainName := options.Name + "." + layerDomain
	domainPath, err := g.projectBuilder.BuildClassLibrary(ctx, domainName, g.fs.Join(root, "src", domainName))
	if err != nil {
		return fmt.Errorf("clean: failed to build domain project: %w", err)
	}

	g.onProgress("Building Application project")
	applicationName := options.Name + "." + layerApplication
	applicationPath, err := g.projectBuilder.BuildClassLibrary(ctx, applicationName, g.fs.Join(root, "src", applicationName))
	if err != nil {
		return fmt.Errorf("clean: failed to build application project: %w", err)
	}

	g.onProgress("Building Infrastructure project")
	infrastructureName := options.Name + "." + layerInfrastructure
	infrastructurePath, err := g.projectBuilder.BuildClassLibrary(ctx, infrastructureName, g.fs.Join(root, "src", infrastructureName))
	if err != nil {
		return fmt.Errorf("clean: failed to build infrastructure project: %w", err)
	}

	g.onProgress(fmt.Sprintf("Building Presentation project (%s)", options.Presentation))
	presentationName := options.Name + "." + string(options.Presentation)
	presentationDir := g.fs.Join(root, "src", presentationName)

	var presentationPath string
	switch options.Presentation {
	case models.PresentationWebApi:
		presentationPath, err = g.projectBuilder.BuildWebApi(ctx, presentationName, presentationDir)
	case models.PresentationWebApp:
		presentationPath, err = g.projectBuilder.BuildMvc(ctx, presentationName, presentationDir)
	default:
		return fmt.Errorf("clean: unsupported presentation type %q", options.Presentation)
	}
	if err != nil {
		return fmt.Errorf("clean: failed to build presentation project: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("clean: cancelled: %w", err)
	}

	g.onProgress("Adding projects to solution")
	for _, projectPath := range []string{domainPath, applicationPath, infrastructurePath, presentationPath} {
		if err := g.referenceBuilder.AddProjectToSolution(ctx, solutionPath, projectPath); err != nil {
			return fmt.Errorf("clean: failed to add %q to solution: %w", projectPath, err)
		}
	}

	g.onProgress("Wiring project references")
	if err := g.referenceBuilder.AddProjectReference(ctx, applicationPath, domainPath); err != nil {
		return fmt.Errorf("clean: failed to reference Domain from Application: %w", err)
	}
	if err := g.referenceBuilder.AddProjectReference(ctx, infrastructurePath, applicationPath); err != nil {
		return fmt.Errorf("clean: failed to reference Application from Infrastructure: %w", err)
	}
	if err := g.referenceBuilder.AddProjectReference(ctx, presentationPath, applicationPath); err != nil {
		return fmt.Errorf("clean: failed to reference Application from Presentation: %w", err)
	}
	if err := g.referenceBuilder.AddProjectReference(ctx, presentationPath, infrastructurePath); err != nil {
		return fmt.Errorf("clean: failed to reference Infrastructure from Presentation: %w", err)
	}

	g.onProgress("Creating conventional folder layout")
	for _, layer := range []string{layerDomain, layerApplication, layerInfrastructure} {
		folders, ok := commonFolders[layer]
		if !ok {
			continue
		}
		projectName := options.Name + "." + layer
		if err := g.folderBuilder.CreateFolders(ctx, g.fs.Join(root, "src", projectName), folders); err != nil {
			return fmt.Errorf("clean: failed to create %s folders: %w", layer, err)
		}
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("clean: cancelled: %w", err)
	}

	g.onProgress(fmt.Sprintf("Installing EF Core packages (%s)", options.Database))
	packages, ok := efCorePackages[options.Database]
	if !ok {
		return fmt.Errorf("clean: unsupported database type %q", options.Database)
	}
	if err := g.packageInstaller.InstallPackages(ctx, infrastructurePath, packages); err != nil {
		return fmt.Errorf("clean: failed to install EF Core packages: %w", err)
	}
	if err := g.packageInstaller.InstallPackage(ctx, infrastructurePath, "Microsoft.EntityFrameworkCore.InMemory", ""); err != nil {
		return fmt.Errorf("clean: failed to install EF Core InMemory package: %w", err)
	}

	g.onProgress("Installing Mediatr packages")
	if err := g.packageInstaller.InstallPackage(ctx, applicationPath, "MediatR", "12.5.0"); err != nil {
		return fmt.Errorf("clean: failed to install Mediatr packages: %w", err)
	}

	g.onProgress("Rendering boilerplate files")
	if err := g.templateRenderer.RenderToFile(
		"infrastructure/AppDbContext.cs.tmpl",
		options,
		g.fs.Join(root, "src", infrastructureName, "Persistence", "AppDbContext.cs"),
	); err != nil {
		return fmt.Errorf("clean: failed to render AppDbContext: %w", err)
	}

	if err := g.templateRenderer.RenderToFile(
		"infrastructure/DependencyInjection.cs.tmpl",
		options,
		g.fs.Join(root, "src", infrastructureName, "DependencyInjection.cs"),
	); err != nil {
		return fmt.Errorf("clean: failed to render DependencyInjection: %w", err)
	}

	if err := g.templateRenderer.RenderToFile(
		"application/DependencyInjection.cs.tmpl",
		options,
		g.fs.Join(root, "src", applicationName, "DependencyInjection.cs"),
	); err != nil {
		return fmt.Errorf("clean: failed to render DependencyInjection: %w", err)
	}

	programTemplate := "api/Program.cs.tmpl"
	if options.Presentation == models.PresentationWebApp {
		programTemplate = "webapp/Program.cs.tmpl"
	}
	if err := g.templateRenderer.RenderToFile(programTemplate, options, g.fs.Join(root, "src", presentationName, "Program.cs")); err != nil {
		return fmt.Errorf("clean: failed to render Program.cs: %w", err)
	}

	if err := g.templateRenderer.RenderToFile(
		"api/appsettings.json.tmpl",
		options,
		g.fs.Join(root, "src", presentationName, "appsettings.json"),
	); err != nil {
		return fmt.Errorf("clean: failed to render appsettings.json: %w", err)
	}

	if err := g.renderCrudExample(root, options); err != nil {
		return err
	}

	if !options.Build {
		g.onProgress("Skipping restore/build (--no-build set)")
		return nil
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("clean: cancelled: %w", err)
	}

	g.onProgress("Restoring NuGet packages")
	if _, err := g.dotnetClient.Restore(ctx, solutionPath); err != nil {
		return fmt.Errorf("clean: failed to restore solution: %w", err)
	}

	g.onProgress("Building solution")
	if _, err := g.dotnetClient.Build(ctx, solutionPath); err != nil {
		return fmt.Errorf("clean: failed to build solution: %w", err)
	}

	g.onProgress("Done")
	return nil
}

// crudExampleTemplates lists the .tmpl -> destination pairs that make up
// the generated WeatherForecast CRUD example, spanning all four layers.
// It exists to prove out the full template pipeline end-to-end (a "hello
// world" that actually compiles and has working endpoints), and is safe to
// delete from a generated project once real entities replace it.
//
// The Presentation-layer controller is handled separately in
// renderCrudExample, since it currently only has a WebApi-flavored
// template (see templates/clean/api/controllers).
func crudExampleTemplates(root string, name string, fs filesystem.FileSystem) []struct {
	template    string
	destination string
} {
	domainName := name + "." + layerDomain
	applicationName := name + "." + layerApplication
	infrastructureName := name + "." + layerInfrastructure
	return []struct {
		template    string
		destination string
	}{
		{"domain/entities/WeatherForecast.cs.tmpl", fs.Join(root, "src", domainName, "Entities", "WeatherForecast.cs")},
		{"domain/interfaces/IWeatherForecastRepository.cs.tmpl", fs.Join(root, "src", domainName, "Interfaces", "IWeatherForecastRepository.cs")},
		{"application/weatherforecasts/dtos/WeatherForecastDto.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "DTOs", "WeatherForecastDto.cs")},
		{"application/weatherforecasts/dtos/CreateWeatherForecastDto.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "DTOs", "CreateWeatherForecastDto.cs")},
		{"application/weatherforecasts/dtos/UpdateWeatherForecastDto.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "DTOs", "UpdateWeatherForecastDto.cs")},
		{"application/weatherforecasts/mappings/WeatherForecastMappingExtensions.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Mappings", "WeatherForecastMappingExtensions.cs")},
		{"application/weatherforecasts/commands/CreateWeatherForecastCommand.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Commands", "CreateWeatherForecastCommand.cs")},
		{"application/weatherforecasts/commands/CreateWeatherForecastCommandHandler.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Commands", "CreateWeatherForecastCommandHandler.cs")},
		{"application/weatherforecasts/commands/UpdateWeatherForecastCommand.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Commands", "UpdateWeatherForecastCommand.cs")},
		{"application/weatherforecasts/commands/UpdateWeatherForecastCommandHandler.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Commands", "UpdateWeatherForecastCommandHandler.cs")},
		{"application/weatherforecasts/commands/DeleteWeatherForecastCommand.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Commands", "DeleteWeatherForecastCommand.cs")},
		{"application/weatherforecasts/commands/DeleteWeatherForecastCommandHandler.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Commands", "DeleteWeatherForecastCommandHandler.cs")},
		{"application/weatherforecasts/queries/GetAllWeatherForecastQuery.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Queries", "GetAllWeatherForecastQuery.cs")},
		{"application/weatherforecasts/queries/GetAllWeatherForecastQueryHandler.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Queries", "GetAllWeatherForecastQueryHandler.cs")},
		{"application/weatherforecasts/queries/GetWeatherForecastByIdQuery.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Queries", "GetWeatherForecastByIdQuery.cs")},
		{"application/weatherforecasts/queries/GetWeatherForecastByIdQueryHandler.cs.tmpl", fs.Join(root, "src", applicationName, "WeatherForecasts", "Queries", "GetWeatherForecastByIdQueryHandler.cs")},
		{"infrastructure/persistence/repositories/WeatherForecastRepository.cs.tmpl", fs.Join(root, "src", infrastructureName, "Persistence", "Repositories", "WeatherForecastRepository.cs")},
	}
}

// renderCrudExample renders the WeatherForecast CRUD vertical slice
// (Domain entity + repository interface, Application DTOs/mapping/service,
// Infrastructure repository, and - for WebApi presentations only - a
// Presentation controller). AppDbContext and DependencyInjection already
// reference WeatherForecast unconditionally (see their .tmpl files), so
// this always runs regardless of --database.
func (g *Generator) renderCrudExample(root string, options models.ProjectOptions) error {
	g.onProgress("Rendering CRUD example (WeatherForecast)")

	for _, file := range crudExampleTemplates(root, options.Name, g.fs) {
		if err := g.templateRenderer.RenderToFile(file.template, options, file.destination); err != nil {
			return fmt.Errorf("clean: failed to render %s: %w", file.template, err)
		}
	}

	presentationName := options.Name + "." + string(options.Presentation)
	switch options.Presentation {
	case models.PresentationWebApi:
		if err := g.templateRenderer.RenderToFile(
			"api/controllers/WeatherForecastsController.cs.tmpl",
			options,
			g.fs.Join(root, "src", presentationName, "Controllers", "WeatherForecastsController.cs"),
		); err != nil {
			return fmt.Errorf("clean: failed to render WeatherForecastsController: %w", err)
		}
	case models.PresentationWebApp:
		// No MVC-flavored controller/view template yet - the Domain,
		// Application, and Infrastructure layers of the example are still
		// generated and functional; only the presentation-layer endpoint
		// is skipped. See README for the current WebApp/MVC limitation.
		g.logger.Info("skipping CRUD example controller: no WebApp/MVC template yet", "presentation", options.Presentation)
	}

	return nil
}
