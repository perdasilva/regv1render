# Validation

## Acceptance Criteria

1. `go build ./...` compiles the entire module including all ported code
2. `go test ./...` passes — all upstream unit and regression tests ported and green
3. `make verify` passes (fmt + vet + lint + test)
4. The public API at `github.com/perdasilva/rv1` exports: `BundleRenderer`, `BundleValidator`, `ResourceGenerator`, `Options`, `CertificateProvider`, and a default `Renderer`
5. All public types and functions have godoc comments
6. No `internal/` package imports the root package
7. `make build` still produces a working rv1 binary

## Test Scenarios

- Run `go test ./...` — all ported upstream unit tests pass
- Run regression tests — all 7 golden-file cases pass (argocd AllNamespaces/SingleNamespace/OwnNamespace, webhook all types, DeploymentConfig options, empty Affinity, empty Affinity subtype)
- Run `make verify` — fmt, vet, lint, test all pass
- Write a small integration snippet that imports `rv1`, creates a default renderer, and calls Render on a minimal bundle — verify it compiles
- Verify no import cycles: `go vet ./...` should catch these

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All upstream unit test cases pass (count matches upstream)
- All 7 regression test cases pass with golden fixtures
- No lint warnings on ported code

## Manual Verification

1. Run `go test -v ./internal/...` and verify test names match upstream
2. Run `go doc github.com/perdasilva/rv1` and verify the public API is clean and documented
3. Inspect `go.mod` — verify all expected dependencies are present
4. Check `internal/` directory structure — verify it mirrors upstream package organization
