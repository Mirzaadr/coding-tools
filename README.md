# DotnetArchitect

`arch` is a cross-platform CLI, written in Go, that generates opinionated
.NET project templates for different software architectures. It shells out
to the `dotnet` CLI to do the actual project creation, and renders
boilerplate files from Go `text/template` templates.

> **Phase 1 (current):** CLI foundation only. `arch create <name>` parses
> and validates arguments, builds a `ProjectOptions` model, resolves the
> `CleanArchitectureGenerator`, and prints the parsed configuration. **No
> .NET projects are generated yet** — that's Phase 2.

## Requirements

- Go 1.22+ to build `arch` itself
- [`dotnet` SDK](https://dotnet.microsoft.com/download) on your `PATH` (only
  needed once generation logic lands in Phase 2)

## Building

All dependencies are vendored under `vendor/`, so the project builds fully
offline:

```bash
go build -o arch ./cmd/arch
```

(Go automatically uses `-mod=vendor` when a `vendor/` directory is present
alongside a consistent `go.mod`/`vendor/modules.txt`, so no extra flags or
network access are required.)

## Usage

```bash
arch create InventorySystem
arch create InventorySystem --presentation WebApi --database PostgreSQL
arch create InventorySystem -p WebApp -db SqlServer -o ./out --no-build
arch create InventorySystem --no-project-folder
```

### `arch create <name>` options

| Flag                   | Short | Values                    | Default             |
| ---------------------- | ----- | -------------------------- | -------------------- |
| `--presentation`       | `-p`  | `WebApi`, `WebApp`          | `WebApi`             |
| `--database`           | `-db` | `PostgreSQL`, `SqlServer`   | `PostgreSQL`         |
| `--output`              | `-o`  | any path                    | current directory     |
| `--no-build`            |       | flag                        | build enabled         |
| `--no-project-folder`   |       | flag                        | project folder created|

> Note: pflag (Cobra's flag library) only supports single-character
> shorthands, so `-db` isn't a "real" pflag shorthand — `arch`'s entrypoint
> (`cmd/arch/main.go`) rewrites `-db` to `--database` before Cobra parses
> the arguments, so it behaves exactly like the spec describes.

Configuration defaults can also be set via a `dotnet-architect.yaml` file
(current directory or `$HOME`) or `DOTNET_ARCHITECT_*` environment
variables, both loaded through Viper — CLI flags always win.

## Architecture

```
cmd/arch/                  Entry point
internal/
  cli/                     Cobra commands + Viper config (only package that imports Cobra)
  architecture/             ArchitectureGenerator interface + registry
    clean/                  Clean Architecture generator
  builders/                 Solution/Project/Reference/Folder/Package/Template builders
  dotnet/                   The only package allowed to shell out to `dotnet`
  filesystem/               Filesystem abstraction (no other package touches os/* directly)
  templates/                Go text/template rendering engine
  models/                   ProjectOptions + enums
  services/                 Business logic (validation, orchestration) - no Cobra dependency
  utils/                    Structured logging (log/slog)
templates/                 .tmpl files consumed by the template engine
test/                      Unit tests, organized by package
```

Design principles:

- **Business logic never imports Cobra.** `internal/services`,
  `internal/architecture`, and `internal/builders` are fully testable
  without any CLI framework in the loop (see `test/`).
- **Only `internal/dotnet` invokes `dotnet`.** Everything else depends on
  the `Dotnet` interface.
- **Only `internal/filesystem` touches `os`/`filepath` for file I/O.**
- **Extensibility:** adding a new architecture (Modular Monolith, Vertical
  Slice, Microservices, ...) means implementing `ArchitectureGenerator` in
  a new `internal/architecture/<name>` package and registering it — no
  existing builder or generator needs to change.

## Roadmap

- **Phase 2:** wire up `CleanArchitectureGenerator.Generate` (already
  implemented, just not yet invoked) to actually produce a Clean
  Architecture solution on disk.
- **Future commands:** `arch add`, `arch remove`, `arch doctor`,
  `arch update`, `arch template`.
- **Future architectures:** Modular Monolith, Vertical Slice, Microservices.

## Testing

```bash
go test ./...
```
