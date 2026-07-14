package architecture_test

import (
	"context"
	"testing"

	"github.com/andre/dotnet-architect/internal/architecture"
	"github.com/andre/dotnet-architect/internal/models"
)

// fakeGenerator is a minimal ArchitectureGenerator test double, proving
// generators can be tested independently of Cobra or the real dotnet CLI.
type fakeGenerator struct {
	name    models.ArchitectureType
	called  bool
	lastOpt models.ProjectOptions
}

func (f *fakeGenerator) Name() models.ArchitectureType {
	return f.name
}

func (f *fakeGenerator) Generate(_ context.Context, options models.ProjectOptions) error {
	f.called = true
	f.lastOpt = options
	return nil
}

func TestRegistry_ResolveRegisteredGenerator(t *testing.T) {
	registry := architecture.NewRegistry()
	fake := &fakeGenerator{name: models.ArchitectureClean}
	registry.Register(fake)

	resolved, err := registry.Resolve(models.ArchitectureClean)
	if err != nil {
		t.Fatalf("expected to resolve registered generator, got error: %v", err)
	}

	if resolved.Name() != models.ArchitectureClean {
		t.Errorf("expected resolved generator name %q, got %q", models.ArchitectureClean, resolved.Name())
	}
}

func TestRegistry_ResolveUnknownArchitecture(t *testing.T) {
	registry := architecture.NewRegistry()

	if _, err := registry.Resolve(models.ArchitectureType("microservices")); err == nil {
		t.Fatal("expected error resolving unregistered architecture, got nil")
	}
}
