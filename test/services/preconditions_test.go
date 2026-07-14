package services_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andre/dotnet-architect/internal/architecture"
	"github.com/andre/dotnet-architect/internal/models"
	"github.com/andre/dotnet-architect/internal/services"
	"github.com/andre/dotnet-architect/internal/utils"
)

func newTestService() *services.CreateProjectService {
	return services.NewCreateProjectService(
		services.NewProjectOptionsValidator(),
		architecture.NewRegistry(),
		utils.NewSilentLogger(),
	)
}

// Note: these tests assume `dotnet` is NOT necessarily on PATH (CI/sandbox
// environments may not have the SDK installed). Where that matters, we
// only assert on the directory-emptiness behavior, which is checked before
// the dotnet lookup in some flows and independent of it in others.

func TestCheckPreconditions_MissingDirectoryIsFine(t *testing.T) {
	svc := newTestService()
	base := t.TempDir()

	options := models.ProjectOptions{
		Name:                "NewProject",
		OutputPath:          base,
		CreateProjectFolder: true, // root = base/NewProject, which doesn't exist yet
	}

	err := svc.CheckPreconditions(options, false)
	if err != nil && !isDotnetMissing(err) {
		t.Fatalf("expected no error (or only a dotnet-not-found error), got: %v", err)
	}
}

func TestCheckPreconditions_RejectsNonEmptyDirWithoutForce(t *testing.T) {
	svc := newTestService()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	options := models.ProjectOptions{
		Name:                "NewProject",
		OutputPath:          dir,
		CreateProjectFolder: false, // root = dir itself, which is non-empty
	}

	err := svc.CheckPreconditions(options, false)
	if err == nil {
		t.Fatal("expected an error for a non-empty output directory without --force")
	}
	if isDotnetMissing(err) {
		t.Skip("dotnet SDK not installed in this environment; can't distinguish the two failure reasons")
	}
}

func TestCheckPreconditions_AllowsNonEmptyDirWithForce(t *testing.T) {
	svc := newTestService()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to seed directory: %v", err)
	}

	options := models.ProjectOptions{
		Name:                "NewProject",
		OutputPath:          dir,
		CreateProjectFolder: false,
	}

	err := svc.CheckPreconditions(options, true)
	if err != nil && !isDotnetMissing(err) {
		t.Fatalf("expected --force to bypass the non-empty-directory check, got: %v", err)
	}
}

// isDotnetMissing reports whether err is the "dotnet SDK not found" failure,
// which is environment-dependent and not what these tests are targeting.
func isDotnetMissing(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		(contains(err.Error(), "dotnet SDK not found"))
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
