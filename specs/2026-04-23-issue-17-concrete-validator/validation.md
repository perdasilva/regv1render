# Validation

## Acceptance Criteria

1. `BundleValidator` is now an interface (not a function-slice type)
2. `Validator` struct exists in `internal/render/registryv1/validator/` with a `Validate()` method
3. `ValidationError` type is exported from the root package
4. `errors.As(err, &ve)` works on validation errors returned by `Validate()`
5. No exported `Check*` functions remain in the validator package
6. `Test_BundleValidatorHasAllValidationFns` guard test is removed
7. All 13 checks still run — same bundles pass/fail identically
8. `make verify` passes
9. All regression tests pass without golden file changes

## Test Scenarios

- Call `Validate()` on a valid bundle — verify no error
- Call `Validate()` on a bundle with duplicate deployment names — verify `ValidationError{Check: "DeploymentSpecUniqueness"}` in result
- Call `Validate()` on a bundle with missing owned CRD — verify `ValidationError{Check: "OwnedCRDExistence"}` in result
- Call `Validate()` with multiple validation failures — verify all `ValidationError` values are present
- Use `errors.As` to extract `ValidationError` — verify it works
- Use `errors.Unwrap` on `ValidationError` — verify underlying error is accessible
- Render a bundle via `DefaultRenderer` — verify validation still runs (invalid bundle still fails)

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All regression tests pass unchanged
- No exported `Check*` functions: `go doc .../validators | grep -c "^func Check"` returns 0

## Manual Verification

1. `go doc github.com/perdasilva/rv1 ValidationError` — verify type is documented
2. `go doc github.com/perdasilva/rv1` — verify `BundleValidator` is gone
3. Run `go test -v ./internal/render/registryv1/validator/` — verify tests pass and use `Validate()`
