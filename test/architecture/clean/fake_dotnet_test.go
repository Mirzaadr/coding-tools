package clean_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/andre/dotnet-architect/internal/dotnet"
)

// recordedCall captures one invocation on the fake Dotnet client, for
// asserting call order and arguments without ever shelling out to a real
// dotnet binary.
type recordedCall struct {
	Op   string
	Args []string
}

// fakeDotnet is an in-memory dotnet.Dotnet test double. It records every
// call it receives and can be configured to fail on a specific operation,
// to exercise error paths (e.g. rollback behavior in services.Generate)
// deterministically.
type fakeDotnet struct {
	mu       sync.Mutex
	calls    []recordedCall
	failOp   string // if set, the next call to this Op returns an error
	failOnce bool   // when true, only the first matching call fails
	failed   bool
}

func newFakeDotnet() *fakeDotnet {
	return &fakeDotnet{}
}

// failOn configures the fake to return an error the next time op is
// invoked (e.g. "CreateWebApi").
func (f *fakeDotnet) failOn(op string) *fakeDotnet {
	f.failOp = op
	f.failOnce = true
	return f
}

func (f *fakeDotnet) record(op string, args ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, recordedCall{Op: op, Args: args})

	if f.failOp == op && f.failOnce && !f.failed {
		f.failed = true
		return fmt.Errorf("fakeDotnet: simulated failure on %s", op)
	}

	return nil
}

func (f *fakeDotnet) callOps() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	ops := make([]string, len(f.calls))
	for i, c := range f.calls {
		ops[i] = c.Op
	}
	return ops
}

func (f *fakeDotnet) CreateSolution(_ context.Context, name string, outputPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("CreateSolution", name, outputPath)
}

func (f *fakeDotnet) CreateClassLibrary(_ context.Context, name string, outputPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("CreateClassLibrary", name, outputPath)
}

func (f *fakeDotnet) CreateWebApi(_ context.Context, name string, outputPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("CreateWebApi", name, outputPath)
}

func (f *fakeDotnet) CreateMvc(_ context.Context, name string, outputPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("CreateMvc", name, outputPath)
}

func (f *fakeDotnet) AddProjectToSolution(_ context.Context, solutionPath string, projectPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("AddProjectToSolution", solutionPath, projectPath)
}

func (f *fakeDotnet) AddReference(_ context.Context, projectPath string, referencedProjectPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("AddReference", projectPath, referencedProjectPath)
}

func (f *fakeDotnet) AddPackage(_ context.Context, projectPath string, packageName string, version string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("AddPackage", projectPath, packageName, version)
}

func (f *fakeDotnet) Build(_ context.Context, targetPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("Build", targetPath)
}

func (f *fakeDotnet) Restore(_ context.Context, targetPath string) (dotnet.Result, error) {
	return dotnet.Result{}, f.record("Restore", targetPath)
}
