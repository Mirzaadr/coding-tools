// Package services hosts business logic that coordinates models,
// architecture generators, and other collaborators on behalf of the CLI
// layer. Nothing in here imports Cobra.
package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andre/dotnet-architect/internal/models"
)

// projectNameRegex mirrors the identifier rules `dotnet new` itself
// enforces reasonably closely: must start with a letter or underscore and
// contain only letters, digits, underscores, or dots.
var projectNameRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)

// ProjectOptionsValidator validates a fully-populated ProjectOptions before
// any generation logic runs.
type ProjectOptionsValidator struct{}

// NewProjectOptionsValidator constructs a ProjectOptionsValidator.
func NewProjectOptionsValidator() *ProjectOptionsValidator {
	return &ProjectOptionsValidator{}
}

// Validate returns an error describing the first validation failure found,
// or nil if options is well-formed.
func (v *ProjectOptionsValidator) Validate(options models.ProjectOptions) error {
	name := strings.TrimSpace(options.Name)
	if name == "" {
		return fmt.Errorf("project name must not be empty")
	}
	if !projectNameRegex.MatchString(name) {
		return fmt.Errorf("project name %q is invalid: must start with a letter or underscore and contain only letters, digits, underscores, or dots", name)
	}

	if strings.TrimSpace(options.OutputPath) == "" {
		return fmt.Errorf("output path must not be empty")
	}

	if options.Presentation != models.PresentationWebApi && options.Presentation != models.PresentationWebApp {
		return fmt.Errorf("unsupported presentation type %q", options.Presentation)
	}

	if options.Database != models.DatabasePostgreSQL && options.Database != models.DatabaseSqlServer {
		return fmt.Errorf("unsupported database type %q", options.Database)
	}

	if options.Architecture != models.ArchitectureClean {
		return fmt.Errorf("unsupported architecture %q", options.Architecture)
	}

	return nil
}
