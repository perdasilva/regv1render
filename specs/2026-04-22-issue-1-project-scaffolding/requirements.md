# Requirements

## Functional Requirements

- `go build ./...` compiles the entire module without errors
- `go test ./...` runs (even if no tests exist yet, it should not fail)
- `make build` produces a binary at `bin/rv1`
- `make test` runs all tests via `go test ./...`
- `make lint` runs `golangci-lint run ./...`
- `make fmt` runs `gofmt` and `goimports` and reports any formatting issues
- `make vet` runs `go vet ./...`
- `make tidy` runs `go mod tidy`
- `make verify` runs fmt, vet, lint, and test in sequence, failing on the first error
- `make clean` removes the `bin/` directory
- GitHub Actions CI runs `make verify` and `make build` on push and PR to main

## Non-Functional Requirements

- Makefile should be simple and readable — no complex scripting or generated targets
- golangci-lint config should be minimal and opinionated (enable useful linters, disable noisy ones)
- CI workflow should complete in under 5 minutes for an empty project
- The module must have no external dependencies at this stage (only stdlib)

## Constraints

- Do not add any rendering logic — this epic is purely scaffolding
- Do not add dependencies beyond the Go standard library
- The `cmd/rv1/main.go` should be a minimal placeholder, not a functional CLI
- Follow the directory structure defined in `specs/tech-stack.md` exactly

## Dependencies

- None — this is the first epic
