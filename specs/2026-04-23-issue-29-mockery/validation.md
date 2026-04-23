# Validation

## Acceptance Criteria

1. `make generate` produces a mock for `CertificateProvider` in `internal/render/certproviders/`
2. `internal/render/fake.go` does not exist
3. `FakeCertProvider` does not appear anywhere in the codebase
4. `rendererTestCertProvider` does not appear anywhere in the codebase
5. `internal/render/resourceutil/` does not exist (merged into `renderutil`)
6. `internal/util/` does not exist (flattened to `internal/testutil/`)
7. `internal/render/util.go` does not exist (callers use `renderutil` directly)
8. `CertProvisionerFor` is unexported
9. `make verify` passes
10. Test count is unchanged

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make generate` runs without errors
- `git diff` after `make generate` shows no changes (mock is committed)
- No import cycles

## Manual Verification

1. `make generate && git diff` — no changes
2. `go vet ./...` — no issues
