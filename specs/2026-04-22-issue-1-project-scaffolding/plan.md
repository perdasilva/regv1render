# Project Scaffolding

Set up the foundational project structure, build system, and CI pipeline so that all subsequent epics have a working development environment with quality gates in place.

## Task Group 1: Module & placeholder code (small)

Initialize the Go module and create minimal source files so the project compiles.

- Initialize `go.mod` with module path `github.com/perdasilva/regv1render` and Go 1.24
- Create `doc.go` at the repo root with a `package regv1render` declaration and a package-level doc comment
- Create `cmd/rv1/main.go` with a minimal `main` package that prints a placeholder message
- Run `go mod tidy` to ensure the module is valid

## Task Group 2: Makefile (small)

Create a Makefile with all build, test, and quality targets.

- Create `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `vet`, `tidy`, `verify`, `clean`
- `build` compiles `cmd/rv1/` and outputs the binary to `bin/rv1`
- `verify` runs `fmt`, `vet`, `lint`, `test` in sequence
- `clean` removes the `bin/` directory
- Verify `make build` produces a working binary

## Task Group 3: Linting & formatting config (small)

Configure golangci-lint and formatting tools.

- Create `.golangci.yml` with a sensible default configuration for a Go library
- Verify `make lint` passes with no issues
- Verify `make fmt` produces no diff

## Task Group 4: CI pipeline (small)

Set up GitHub Actions for continuous integration.

- Create `.github/workflows/ci.yml` with a workflow triggered on push and pull request to main
- Steps: checkout, setup Go 1.24, `make verify`, `make build`
- Ensure the workflow uses a matrix or single job as appropriate for this project size

## Task Group 5: Directory structure & .gitignore (small)

Create the remaining directory stubs and ignore rules.

- Create `internal/` directory with a `.gitkeep` file
- Create `testdata/` directory with a `.gitkeep` file
- Create `.gitignore` covering Go build artifacts (`bin/`, `*.exe`, etc.), editor files, and OS files
- Verify the final directory structure matches `specs/tech-stack.md`
