# Validation

## Acceptance Criteria

1. `go test -run Example ./` passes — all testable examples compile and run
2. `go doc github.com/perdasilva/rv1` shows examples in output
3. All public types and functions have godoc comments
4. CONTRIBUTING.md exists with prerequisites, build, test, conventions, and PR process
5. README has upstream relationship section
6. README has OLMv0 compatibility section
7. `make verify` passes
8. `make build` still produces a working rv1 binary

## Test Scenarios

- Run `go test -run Example -v ./` — verify all examples pass and produce expected output
- Run `go doc github.com/perdasilva/rv1 Render` — verify example appears
- Run `go doc github.com/perdasilva/rv1 FromFS` — verify example appears
- Read CONTRIBUTING.md — verify it covers clone, build, test, branch, commit, PR
- Read README — verify upstream relationship section is present and accurate
- Read README — verify OLMv0 section explains WithProvidedAPIsClusterRoles

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All existing tests pass unchanged
- New examples pass as part of `go test`

## Manual Verification

1. Run `go doc github.com/perdasilva/rv1` — review the full package documentation
2. Read README.md end-to-end — verify it tells a coherent story
3. Read CONTRIBUTING.md — verify a new contributor could follow the instructions
