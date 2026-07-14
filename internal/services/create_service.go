package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andre/dotnet-architect/internal/architecture"
	"github.com/andre/dotnet-architect/internal/models"
)

// CreateProjectService coordinates the `arch create` workflow: validating
// options and resolving the requested ArchitectureGenerator. It contains no
// Cobra dependency, so it can be tested and reused independently of the CLI.
type CreateProjectService struct {
	validator *ProjectOptionsValidator
	registry  *architecture.Registry
	logger    *slog.Logger
}

// NewCreateProjectService constructs a CreateProjectService.
func NewCreateProjectService(validator *ProjectOptionsValidator, registry *architecture.Registry, logger *slog.Logger) *CreateProjectService {
	return &CreateProjectService{validator: validator, registry: registry, logger: logger}
}

// PrepareResult is returned by Prepare and contains everything the CLI
// layer needs to report back to the user.
type PrepareResult struct {
	Options   models.ProjectOptions
	Generator architecture.ArchitectureGenerator
}

// Prepare validates options and resolves the matching ArchitectureGenerator.
// It does NOT invoke Generator.Generate - actual project generation is left
// to a later phase, so this is safe to call as a dry-run / preview step.
func (s *CreateProjectService) Prepare(options models.ProjectOptions) (PrepareResult, error) {
	if err := s.validator.Validate(options); err != nil {
		return PrepareResult{}, fmt.Errorf("invalid options: %w", err)
	}

	generator, err := s.registry.Resolve(options.Architecture)
	if err != nil {
		return PrepareResult{}, fmt.Errorf("failed to resolve generator: %w", err)
	}

	s.logger.Debug("resolved architecture generator",
		"architecture", options.Architecture,
		"project", options.Name,
	)

	return PrepareResult{Options: options, Generator: generator}, nil
}

// GenerateSettings controls behavior that isn't part of the generated
// project itself (safety checks, cleanup policy).
type GenerateSettings struct {
	// Force allows generating into a non-empty output directory.
	Force bool
	// KeepPartial, when true, leaves partially-generated output on disk if
	// generation fails partway through. When false (default), the project
	// root is removed on failure IF this call is what created it.
	KeepPartial bool
}

// Generate runs preconditions, then invokes the resolved generator. It is
// the only method in this service that touches disk or shells out.
func (s *CreateProjectService) Generate(ctx context.Context, result PrepareResult, settings GenerateSettings) error {
	root := projectRoot(result.Options)

	rootPreexisted, err := s.checkPreconditions(result.Options, root, settings.Force)
	if err != nil {
		return fmt.Errorf("preconditions failed: %w", err)
	}

	s.logger.Info("generating project",
		"name", result.Options.Name,
		"architecture", result.Options.Architecture,
		"root", root,
	)

	if err := result.Generator.Generate(ctx, result.Options); err != nil {
		if !settings.KeepPartial && !rootPreexisted {
			s.logger.Warn("generation failed, removing partially generated output", "root", root)
			if removeErr := os.RemoveAll(root); removeErr != nil {
				s.logger.Error("failed to clean up partial output", "root", root, "error", removeErr)
			}
		} else {
			s.logger.Warn("generation failed, leaving partial output in place for inspection", "root", root)
		}
		return fmt.Errorf("generation failed: %w", err)
	}

	return nil
}

// CheckPreconditions verifies it's safe to generate for the given options
// without actually generating anything. Exported so it can be reused (e.g.
// by a future `arch doctor` command) and tested directly.
func (s *CreateProjectService) CheckPreconditions(options models.ProjectOptions, force bool) error {
	_, err := s.checkPreconditions(options, projectRoot(options), force)
	return err
}

// checkPreconditions returns whether root already existed before this call,
// which Generate uses to decide whether cleanup-on-failure is appropriate.
func (s *CreateProjectService) checkPreconditions(options models.ProjectOptions, root string, force bool) (bool, error) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		return false, fmt.Errorf("dotnet SDK not found on PATH: install it from https://dotnet.microsoft.com/download")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect output directory %q: %w", root, err)
	}

	if len(entries) > 0 && !force {
		return true, fmt.Errorf("output directory %q is not empty (use --force to generate anyway)", root)
	}

	return true, nil
}

// projectRoot resolves the actual directory generation will target, taking
// options.CreateProjectFolder into account. This must match the logic
// architecture generators use internally so preconditions and cleanup point
// at the right place.
func projectRoot(options models.ProjectOptions) string {
	if options.CreateProjectFolder {
		return filepath.Join(options.OutputPath, options.Name)
	}
	return options.OutputPath
}

// Summary renders a human-readable description of the parsed configuration.
// Building the string here (rather than in the CLI layer) keeps formatting
// logic testable; the CLI layer is only responsible for printing it.
func Summary(options models.ProjectOptions) string {
	projectFolder := "no (files placed directly in output path)"
	if options.CreateProjectFolder {
		projectFolder = "yes"
	}

	build := "no (--no-build)"
	if options.Build {
		build = "yes"
	}

	return fmt.Sprintf(
		"Parsed configuration:\n"+
			"  Name:                 %s\n"+
			"  Architecture:         %s\n"+
			"  Presentation:         %s\n"+
			"  Database:             %s\n"+
			"  Output path:          %s\n"+
			"  Create project folder: %s\n"+
			"  Restore/Build after generation: %s\n",
		options.Name,
		options.Architecture,
		options.Presentation,
		options.Database,
		options.OutputPath,
		projectFolder,
		build,
	)
}
