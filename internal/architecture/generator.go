// Package architecture defines the ArchitectureGenerator contract and a
// registry for resolving generators by name. Individual architectures
// (clean, modular, vertical, ...) live in their own subpackages and are
// registered here without this package needing to know their internals.
package architecture

import (
	"context"
	"fmt"

	"github.com/andre/dotnet-architect/internal/models"
)

// ArchitectureGenerator generates a full project layout for a specific
// architecture style. Adding a new architecture means implementing this
// interface in a new subpackage and registering it - no existing generator
// or builder needs to change (Open/Closed Principle).
type ArchitectureGenerator interface {
	// Generate produces the project on disk according to options.
	Generate(ctx context.Context, options models.ProjectOptions) error

	// Name returns the architecture identifier this generator handles,
	// e.g. "clean".
	Name() models.ArchitectureType
}

// ProgressFunc is an optional callback generators can invoke to report
// human-readable progress steps ("Creating solution", "Building Domain
// project", ...). It's a plain func type (not an interface) so callers can
// pass a closure, and generators can default to a no-op when nil.
type ProgressFunc func(step string)

// Registry resolves an ArchitectureGenerator by its ArchitectureType.
type Registry struct {
	generators map[models.ArchitectureType]ArchitectureGenerator
}

// NewRegistry constructs an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[models.ArchitectureType]ArchitectureGenerator),
	}
}

// Register adds a generator to the registry, keyed by its own Name().
func (r *Registry) Register(generator ArchitectureGenerator) {
	r.generators[generator.Name()] = generator
}

// Resolve looks up the generator registered for the given architecture
// type. It returns an error if no generator has been registered for it.
func (r *Registry) Resolve(architecture models.ArchitectureType) (ArchitectureGenerator, error) {
	generator, ok := r.generators[architecture]
	if !ok {
		return nil, fmt.Errorf("architecture: no generator registered for architecture %q", architecture)
	}
	return generator, nil
}
