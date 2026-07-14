package services_test

import (
	"testing"

	"github.com/andre/dotnet-architect/internal/models"
	"github.com/andre/dotnet-architect/internal/services"
)

func validOptions() models.ProjectOptions {
	return models.ProjectOptions{
		Name:                "InventorySystem",
		Architecture:        models.ArchitectureClean,
		OutputPath:          "/tmp/out",
		Presentation:        models.PresentationWebApi,
		Database:            models.DatabasePostgreSQL,
		Build:               true,
		CreateProjectFolder: true,
	}
}

func TestValidator_AcceptsValidOptions(t *testing.T) {
	v := services.NewProjectOptionsValidator()

	if err := v.Validate(validOptions()); err != nil {
		t.Fatalf("expected valid options to pass, got error: %v", err)
	}
}

func TestValidator_RejectsEmptyName(t *testing.T) {
	v := services.NewProjectOptionsValidator()
	opts := validOptions()
	opts.Name = "   "

	if err := v.Validate(opts); err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestValidator_RejectsInvalidNameCharacters(t *testing.T) {
	v := services.NewProjectOptionsValidator()
	opts := validOptions()
	opts.Name = "1 Invalid Name!"

	if err := v.Validate(opts); err == nil {
		t.Fatal("expected error for invalid name characters, got nil")
	}
}

func TestValidator_RejectsEmptyOutputPath(t *testing.T) {
	v := services.NewProjectOptionsValidator()
	opts := validOptions()
	opts.OutputPath = ""

	if err := v.Validate(opts); err == nil {
		t.Fatal("expected error for empty output path, got nil")
	}
}

func TestValidator_RejectsUnsupportedPresentation(t *testing.T) {
	v := services.NewProjectOptionsValidator()
	opts := validOptions()
	opts.Presentation = models.PresentationType("Desktop")

	if err := v.Validate(opts); err == nil {
		t.Fatal("expected error for unsupported presentation, got nil")
	}
}

func TestValidator_RejectsUnsupportedDatabase(t *testing.T) {
	v := services.NewProjectOptionsValidator()
	opts := validOptions()
	opts.Database = models.DatabaseType("MongoDB")

	if err := v.Validate(opts); err == nil {
		t.Fatal("expected error for unsupported database, got nil")
	}
}

func TestValidator_RejectsUnsupportedArchitecture(t *testing.T) {
	v := services.NewProjectOptionsValidator()
	opts := validOptions()
	opts.Architecture = models.ArchitectureType("microservices")

	if err := v.Validate(opts); err == nil {
		t.Fatal("expected error for unsupported architecture, got nil")
	}
}
