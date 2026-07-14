package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/andre/dotnet-architect/internal/architecture"
	"github.com/andre/dotnet-architect/internal/architecture/clean"
	"github.com/andre/dotnet-architect/internal/builders"
	"github.com/andre/dotnet-architect/internal/cli/config"
	"github.com/andre/dotnet-architect/internal/dotnet"
	"github.com/andre/dotnet-architect/internal/filesystem"
	"github.com/andre/dotnet-architect/internal/models"
	"github.com/andre/dotnet-architect/internal/services"
	"github.com/andre/dotnet-architect/internal/templates"
)

// createFlags holds the raw, unvalidated flag values Cobra populates. It is
// translated into a models.ProjectOptions (and validated) by the create
// command's RunE, keeping the parsing/validation boundary explicit.
type createFlags struct {
	presentation    string
	database        string
	output          string
	noBuild         bool
	noProjectFolder bool
	dryRun          bool
	force           bool
	keepPartial     bool
}

// NewCreateCommand builds the `arch create <name>` command. loggerFn is a
// lazily-evaluated accessor because the root command constructs the actual
// *slog.Logger only after flags (like --verbose) are parsed.
func NewCreateCommand(loggerFn func() *slog.Logger, cfg *config.Config) *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new .NET project from an architecture template",
		Long: "Create a new .NET project from an architecture template.\n\n" +
			"Example:\n  arch create InventorySystem --presentation WebApi --database PostgreSQL",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(loggerFn(), args[0], flags)
		},
	}

	cmd.Flags().StringVarP(&flags.presentation, "presentation", "p", cfg.String("presentation"), "presentation layer: WebApi or WebApp")
	// pflag only allows single-character shorthands (POSIX/GNU convention),
	// so the "-db" short form from the spec can't be registered as a true
	// pflag shorthand. It's supported anyway via normalizeDBShorthand, which
	// rewrites "-db" to "--database" before Cobra parses args (see main.go).
	cmd.Flags().StringVar(&flags.database, "database", cfg.String("database"), "database provider: PostgreSQL or SqlServer (short form: -db)")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "output directory (default: current directory)")
	cmd.Flags().BoolVar(&flags.noBuild, "no-build", false, "skip dotnet restore/build after generation")
	cmd.Flags().BoolVar(&flags.noProjectFolder, "no-project-folder", false, "generate directly into the output directory instead of a new named subfolder")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "preview the changes without actually creating files")
	cmd.Flags().BoolVar(&flags.force, "force", false, "overwrite existing files")
	cmd.Flags().BoolVar(&flags.keepPartial, "keep-partial", false, "keep partially-generated output on failure")

	return cmd
}

// runCreate translates flags into a models.ProjectOptions, validates it,
// resolves the requested ArchitectureGenerator, and reports the parsed
// configuration back to the user. Per the Phase 1 scope, it does NOT invoke
// generator.Generate - no .NET projects are produced yet.
func runCreate(logger *slog.Logger, name string, flags *createFlags) error {
	presentation, err := models.ParsePresentationType(flags.presentation)
	if err != nil {
		return err
	}

	database, err := models.ParseDatabaseType(flags.database)
	if err != nil {
		return err
	}

	outputPath := flags.output
	if outputPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to resolve current directory: %w", err)
		}
		outputPath = cwd
	}

	options := models.ProjectOptions{
		Name:                name,
		Architecture:        models.ArchitectureClean,
		OutputPath:          outputPath,
		Presentation:        presentation,
		Database:            database,
		Build:               !flags.noBuild,
		CreateProjectFolder: !flags.noProjectFolder,
	}

	createService := buildCreateProjectService(logger)

	result, err := createService.Prepare(options)
	if err != nil {
		return err
	}

	fmt.Println(services.Summary(result.Options))

	if flags.dryRun {
		fmt.Printf("Dry run: generator %q no files will be created.", result.Generator.Name())
		return nil
	}

	// Cancel generation cleanly on Ctrl+C / SIGTERM instead of leaving a
	// dangling dotnet subprocess or getting killed mid-write.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	settings := services.GenerateSettings{
		Force:       flags.force,
		KeepPartial: flags.keepPartial,
	}

	if err := createService.Generate(ctx, result, settings); err != nil {
		return err
	}

	fmt.Println("Project generation completed successfully.")
	return nil
}

// buildCreateProjectService performs the explicit, constructor-based
// dependency wiring for the create workflow: filesystem -> dotnet wrapper
// -> template engine -> builders -> architecture generators -> registry ->
// service. No dependency injection framework is used.
func buildCreateProjectService(logger *slog.Logger) *services.CreateProjectService {
	fs := filesystem.NewOSFileSystem()
	dotnetClient := dotnet.NewCLIDotnet(logger)
	templateEngine := templates.NewEngine(resolveTemplatesRoot(logger))

	solutionBuilder := builders.NewDotnetSolutionBuilder(dotnetClient, logger)
	projectBuilder := builders.NewDotnetProjectBuilder(dotnetClient, fs, logger)
	referenceBuilder := builders.NewDotnetReferenceBuilder(dotnetClient, logger)
	folderBuilder := builders.NewStandardFolderBuilder(fs, logger)
	packageInstaller := builders.NewDotnetPackageInstaller(dotnetClient, logger)
	templateRenderer := builders.NewEngineTemplateRenderer(templateEngine, fs, logger)

	onProgress := architecture.ProgressFunc(func(step string) {
		fmt.Println("→", step)
	})

	cleanGenerator := clean.NewGenerator(
		solutionBuilder,
		projectBuilder,
		referenceBuilder,
		folderBuilder,
		packageInstaller,
		templateRenderer,
		fs,
		dotnetClient,
		logger,
		onProgress,
	)

	registry := architecture.NewRegistry()
	registry.Register(cleanGenerator)

	validator := services.NewProjectOptionsValidator()

	return services.NewCreateProjectService(validator, registry, logger)
}

// resolveTemplatesRoot locates the "templates/clean" directory. It checks,
// in order: next to the running executable (the common case for an
// installed binary), then the current working directory (the common case
// for `go run` during development). Falling back to a relative path keeps
// existing behavior if neither resolves - the eventual "file not found"
// error from the template engine will point at the real problem.
func resolveTemplatesRoot(logger *slog.Logger) string {
	const relative = "templates/clean"

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), relative)
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
	}

	if info, err := os.Stat(relative); err == nil && info.IsDir() {
		return relative
	}

	logger.Warn("could not locate templates/clean next to the executable or in the working directory; falling back to relative path", "path", relative)
	return relative
}
