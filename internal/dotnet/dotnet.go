// Package dotnet is the only package in the application allowed to invoke
// the `dotnet` CLI. Every other package that needs dotnet behavior must
// depend on the Dotnet interface defined here.
package dotnet

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

// Result captures the outcome of a dotnet CLI invocation.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// CommandError is returned when a dotnet invocation fails. It carries the
// captured output so callers can present meaningful diagnostics.
type CommandError struct {
	Args     []string
	ExitCode int
	Stderr   string
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("dotnet %v failed with exit code %d: %s", e.Args, e.ExitCode, e.Stderr)
}

// Dotnet abstracts every `dotnet` CLI operation the application needs.
// Implementations must never be called directly with os/exec from outside
// this package.
type Dotnet interface {
	CreateSolution(ctx context.Context, name string, outputPath string) (Result, error)
	CreateClassLibrary(ctx context.Context, name string, outputPath string) (Result, error)
	CreateWebApi(ctx context.Context, name string, outputPath string) (Result, error)
	CreateMvc(ctx context.Context, name string, outputPath string) (Result, error)
	AddProjectToSolution(ctx context.Context, solutionPath string, projectPath string) (Result, error)
	AddReference(ctx context.Context, projectPath string, referencedProjectPath string) (Result, error)
	AddPackage(ctx context.Context, projectPath string, packageName string, version string) (Result, error)
	Build(ctx context.Context, targetPath string) (Result, error)
	Restore(ctx context.Context, targetPath string) (Result, error)
}

// CLIDotnet is the production Dotnet implementation, shelling out to the
// `dotnet` executable via os/exec.
type CLIDotnet struct {
	logger *slog.Logger
}

// NewCLIDotnet constructs a CLIDotnet wrapper.
func NewCLIDotnet(logger *slog.Logger) *CLIDotnet {
	return &CLIDotnet{logger: logger}
}

func (d *CLIDotnet) run(ctx context.Context, args ...string) (Result, error) {
	d.logger.Debug("running dotnet command", "args", args)

	cmd := exec.CommandContext(ctx, "dotnet", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		return result, &CommandError{
			Args:     args,
			ExitCode: result.ExitCode,
			Stderr:   result.Stderr,
		}
	}

	result.ExitCode = 0
	return result, nil
}

func (d *CLIDotnet) CreateSolution(ctx context.Context, name string, outputPath string) (Result, error) {
	return d.run(ctx, "new", "sln", "-n", name, "-o", outputPath)
}

func (d *CLIDotnet) CreateClassLibrary(ctx context.Context, name string, outputPath string) (Result, error) {
	return d.run(ctx, "new", "classlib", "-n", name, "-o", outputPath)
}

func (d *CLIDotnet) CreateWebApi(ctx context.Context, name string, outputPath string) (Result, error) {
	return d.run(ctx, "new", "webapi", "-n", name, "-o", outputPath)
}

func (d *CLIDotnet) CreateMvc(ctx context.Context, name string, outputPath string) (Result, error) {
	return d.run(ctx, "new", "mvc", "-n", name, "-o", outputPath)
}

func (d *CLIDotnet) AddProjectToSolution(ctx context.Context, solutionPath string, projectPath string) (Result, error) {
	return d.run(ctx, "sln", solutionPath, "add", projectPath)
}

func (d *CLIDotnet) AddReference(ctx context.Context, projectPath string, referencedProjectPath string) (Result, error) {
	return d.run(ctx, "add", projectPath, "reference", referencedProjectPath)
}

func (d *CLIDotnet) AddPackage(ctx context.Context, projectPath string, packageName string, version string) (Result, error) {
	args := []string{"add", projectPath, "package", packageName}
	if version != "" {
		args = append(args, "-v", version)
	}
	return d.run(ctx, args...)
}

func (d *CLIDotnet) Build(ctx context.Context, targetPath string) (Result, error) {
	return d.run(ctx, "build", targetPath)
}

func (d *CLIDotnet) Restore(ctx context.Context, targetPath string) (Result, error) {
	return d.run(ctx, "restore", targetPath)
}
