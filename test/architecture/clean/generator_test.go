package clean_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andre/dotnet-architect/internal/architecture/clean"
	"github.com/andre/dotnet-architect/internal/builders"
	"github.com/andre/dotnet-architect/internal/filesystem"
	"github.com/andre/dotnet-architect/internal/models"
	"github.com/andre/dotnet-architect/internal/templates"
	"github.com/andre/dotnet-architect/internal/utils"
)

// newTestTemplatesDir writes minimal .tmpl files at the exact relative
// paths clean.Generator.Generate renders, so tests don't depend on the
// real repository templates/ directory or its content.
func newTestTemplatesDir(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	files := map[string]string{
		"infrastructure/AppDbContext.cs.tmpl":        "// AppDbContext for {{ .Name }} ({{ .Database }})\n",
		"infrastructure/DependencyInjection.cs.tmpl": "// DI for {{ .Name }}\n",
		"api/Program.cs.tmpl":                        "// Program.cs for {{ .Name }} ({{ .Presentation }})\n",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("failed to create template dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write template file: %v", err)
		}
	}

	return root
}

// newTestGenerator wires a clean.Generator using real builders, a real
// filesystem rooted at t.TempDir(), and real template rendering, but a
// fake dotnet.Dotnet client - so tests exercise the full orchestration and
// real file output without ever shelling out to a real `dotnet` binary.
func newTestGenerator(t *testing.T, fake *fakeDotnet) (*clean.Generator, *[]string) {
	t.Helper()

	logger := utils.NewSilentLogger()
	fs := filesystem.NewOSFileSystem()
	templateEngine := templates.NewEngine(newTestTemplatesDir(t))

	solutionBuilder := builders.NewDotnetSolutionBuilder(fake, logger)
	projectBuilder := builders.NewDotnetProjectBuilder(fake, fs, logger)
	referenceBuilder := builders.NewDotnetReferenceBuilder(fake, logger)
	folderBuilder := builders.NewStandardFolderBuilder(fs, logger)
	packageInstaller := builders.NewDotnetPackageInstaller(fake, logger)
	templateRenderer := builders.NewEngineTemplateRenderer(templateEngine, fs, logger)

	progressSteps := &[]string{}
	onProgress := func(step string) { *progressSteps = append(*progressSteps, step) }

	gen := clean.NewGenerator(
		solutionBuilder,
		projectBuilder,
		referenceBuilder,
		folderBuilder,
		packageInstaller,
		templateRenderer,
		fs,
		fake,
		logger,
		onProgress,
	)

	return gen, progressSteps
}

func baseOptions(t *testing.T) models.ProjectOptions {
	t.Helper()
	return models.ProjectOptions{
		Name:                "InventorySystem",
		Architecture:        models.ArchitectureClean,
		OutputPath:          t.TempDir(),
		Presentation:        models.PresentationWebApi,
		Database:            models.DatabasePostgreSQL,
		Build:               true,
		CreateProjectFolder: true,
	}
}

func TestGenerate_HappyPath_CallsDotnetInExpectedOrder(t *testing.T) {
	fake := newFakeDotnet()
	gen, progressPtr := newTestGenerator(t, fake)
	options := baseOptions(t)

	if err := gen.Generate(context.Background(), options); err != nil {
		t.Fatalf("expected Generate to succeed, got error: %v", err)
	}

	got := fake.callOps()
	want := []string{
		"CreateSolution",
		"CreateClassLibrary", // Domain
		"CreateClassLibrary", // Application
		"CreateClassLibrary", // Infrastructure
		"CreateWebApi",       // Presentation
		"AddProjectToSolution", "AddProjectToSolution", "AddProjectToSolution", "AddProjectToSolution",
		"AddReference", "AddReference", "AddReference", "AddReference",
		"AddPackage", // Npgsql
		"Restore",
		"Build",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected dotnet call sequence:\n got:  %v\n want: %v", got, want)
	}

	progress := *progressPtr
	if len(progress) == 0 {
		t.Error("expected progress callback to be invoked at least once")
	}
	if progress[len(progress)-1] != "Done" {
		t.Errorf("expected last progress step to be %q, got %q", "Done", progress[len(progress)-1])
	}
}

func TestGenerate_WebApp_UsesMvcNotWebApi(t *testing.T) {
	fake := newFakeDotnet()
	gen, _ := newTestGenerator(t, fake)
	options := baseOptions(t)
	options.Presentation = models.PresentationWebApp

	if err := gen.Generate(context.Background(), options); err != nil {
		t.Fatalf("expected Generate to succeed, got error: %v", err)
	}

	for _, op := range fake.callOps() {
		if op == "CreateWebApi" {
			t.Error("expected CreateMvc for WebApp presentation, but CreateWebApi was called")
		}
	}
}

func TestGenerate_SqlServer_InstallsSqlServerPackage(t *testing.T) {
	fake := newFakeDotnet()
	gen, _ := newTestGenerator(t, fake)
	options := baseOptions(t)
	options.Database = models.DatabaseSqlServer

	if err := gen.Generate(context.Background(), options); err != nil {
		t.Fatalf("expected Generate to succeed, got error: %v", err)
	}

	found := false
	for _, c := range fake.calls {
		if c.Op == "AddPackage" && len(c.Args) > 1 && c.Args[1] == "Microsoft.EntityFrameworkCore.SqlServer" {
			found = true
		}
	}
	if !found {
		t.Error("expected Microsoft.EntityFrameworkCore.SqlServer package to be installed for SqlServer database")
	}
}

func TestGenerate_NoBuild_SkipsRestoreAndBuild(t *testing.T) {
	fake := newFakeDotnet()
	gen, progressPtr := newTestGenerator(t, fake)
	options := baseOptions(t)
	options.Build = false

	if err := gen.Generate(context.Background(), options); err != nil {
		t.Fatalf("expected Generate to succeed, got error: %v", err)
	}

	for _, op := range fake.callOps() {
		if op == "Restore" || op == "Build" {
			t.Errorf("expected no %s call when Build=false", op)
		}
	}

	last := (*progressPtr)[len(*progressPtr)-1]
	if last != "Skipping restore/build (--no-build set)" {
		t.Errorf("expected final progress step to note skipped build, got %q", last)
	}
}

func TestGenerate_FailurePropagatesWithContext(t *testing.T) {
	fake := newFakeDotnet().failOn("CreateWebApi")
	gen, _ := newTestGenerator(t, fake)
	options := baseOptions(t)

	err := gen.Generate(context.Background(), options)
	if err == nil {
		t.Fatal("expected Generate to return an error when dotnet fails, got nil")
	}
	if !contains(err.Error(), "presentation project") {
		t.Errorf("expected error to mention the failing step, got: %v", err)
	}

	// Domain/Application/Infrastructure should have been created before the
	// simulated failure on the Presentation project.
	got := fake.callOps()
	want := []string{"CreateSolution", "CreateClassLibrary", "CreateClassLibrary", "CreateClassLibrary", "CreateWebApi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected dotnet call sequence before failure:\n got:  %v\n want: %v", got, want)
	}

	// Nothing past the failure point should have run.
	for _, op := range got {
		if op == "AddProjectToSolution" || op == "Restore" || op == "Build" {
			t.Errorf("did not expect %s to run after CreateWebApi failed", op)
		}
	}
}

func TestGenerate_RendersTemplateFilesToExpectedPaths(t *testing.T) {
	fake := newFakeDotnet()
	gen, _ := newTestGenerator(t, fake)
	options := baseOptions(t)

	if err := gen.Generate(context.Background(), options); err != nil {
		t.Fatalf("expected Generate to succeed, got error: %v", err)
	}

	root := filepath.Join(options.OutputPath, options.Name)
	expectedFiles := []string{
		filepath.Join(root, "Infrastructure", "Persistence", "AppDbContext.cs"),
		filepath.Join(root, "Infrastructure", "DependencyInjection", "DependencyInjection.cs"),
		filepath.Join(root, "Presentation", "Program.cs"),
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected rendered file to exist at %q: %v", f, err)
		}
	}

	content, err := os.ReadFile(filepath.Join(root, "Infrastructure", "Persistence", "AppDbContext.cs"))
	if err != nil {
		t.Fatalf("failed to read rendered AppDbContext.cs: %v", err)
	}
	if !contains(string(content), "InventorySystem") || !contains(string(content), "PostgreSQL") {
		t.Errorf("expected rendered template to substitute Name and Database, got: %q", content)
	}
}

func TestGenerate_CreatesConventionalFolders(t *testing.T) {
	fake := newFakeDotnet()
	gen, _ := newTestGenerator(t, fake)
	options := baseOptions(t)

	if err := gen.Generate(context.Background(), options); err != nil {
		t.Fatalf("expected Generate to succeed, got error: %v", err)
	}

	root := filepath.Join(options.OutputPath, options.Name)
	expectedDirs := []string{
		filepath.Join(root, "Domain", "Entities"),
		filepath.Join(root, "Application", "DTOs"),
		filepath.Join(root, "Infrastructure", "Services"),
	}

	for _, d := range expectedDirs {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("expected conventional folder to exist at %q: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %q to be a directory", d)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
