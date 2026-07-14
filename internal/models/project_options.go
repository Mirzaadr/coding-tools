package models

// ArchitectureType represents which architecture generator should be used.
// This is intentionally a simple string type so new architectures can be
// registered without changing this package.
type ArchitectureType string

const (
	ArchitectureClean ArchitectureType = "clean"
)

// ProjectOptions is the fully-parsed and validated representation of a
// `arch create` invocation. It is the single input contract that every
// ArchitectureGenerator implementation consumes.
type ProjectOptions struct {
	// Name is the name of the project/solution to generate (e.g. "InventorySystem").
	Name string

	// Architecture selects which ArchitectureGenerator will be resolved.
	Architecture ArchitectureType

	// OutputPath is the directory the project should be generated into.
	OutputPath string

	// Presentation selects the presentation layer flavor (WebApi, WebApp).
	Presentation PresentationType

	// Database selects the database provider (PostgreSQL, SqlServer).
	Database DatabaseType

	// Build indicates whether `dotnet restore` and `dotnet build` should run
	// after generation. Corresponds to the inverse of --no-build.
	Build bool

	// CreateProjectFolder indicates whether a dedicated folder named after
	// the project should be created under OutputPath. Corresponds to the
	// inverse of --no-project-folder.
	CreateProjectFolder bool
}
