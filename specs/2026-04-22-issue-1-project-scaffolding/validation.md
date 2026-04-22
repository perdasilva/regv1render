# Validation

## Acceptance Criteria

1. `make verify` passes (fmt + vet + lint + test all succeed)
2. `make build` produces a `bin/rv1` binary that runs without error
3. `go mod tidy` produces no diff (module is clean)
4. GitHub Actions CI workflow runs successfully on push
5. Directory structure matches `specs/tech-stack.md`

## Test Scenarios

- Run `make verify` from a clean checkout — all targets pass
- Run `make build` — binary appears at `bin/rv1`
- Run `bin/rv1` — prints a placeholder message and exits cleanly
- Run `make clean` — `bin/` directory is removed
- Run `go mod tidy && git diff go.mod go.sum` — no changes
- Push to remote — GitHub Actions CI triggers and passes

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- `go mod tidy` is clean
- No external dependencies in `go.mod`

## Manual Verification

1. Clone the repo on a clean machine with Go 1.25
2. Run `make verify` — should pass with no prerequisites beyond Go and golangci-lint
3. Run `make build && bin/rv1` — binary runs
4. Push a commit — verify GitHub Actions CI completes green
