# Refactor Validator to Concrete Type with Structured Errors

Replace the generic `BundleValidator` function-slice pattern with a concrete `Validator` struct that owns all 13 validation checks internally, returns structured `ValidationError` types, and simplifies the public API.

## Task Group 1: Add ValidationError type (small)

Define the structured error type in `internal/render/`.

- Create `ValidationError` struct with `Check` (string) and `Err` (error) fields
- Implement `Error()` and `Unwrap()` methods
- Export `ValidationError` from the public API in `render.go`

## Task Group 2: Refactor validator to concrete struct (medium)

Replace the function-slice pattern with a concrete `Validator` struct.

- Create `Validator` struct in `internal/render/registryv1/validator/`
- Add `Validate(*bundle.RegistryV1) error` method that runs all 13 checks
- Wrap each check's errors in `ValidationError` with the check name
- Unexport the 13 `Check*` functions (rename to lowercase)
- Change `BundleValidator` from function-slice type to interface in `internal/render/render.go`
- Add nil-validator guard in `Render()` to handle zero-value `BundleRenderer{}`

## Task Group 3: Update wiring and public API (small)

Wire the new validator into the renderer and update the public API.

- Update `BundleRenderer` struct to hold `validators.Validator` instead of `BundleValidator`
- Update `registryv1.go` to use the new concrete `Validator`
- Update `render.go` public API — `BundleValidator` now aliases the interface, export `ValidationError`
- Update godoc comments

## Task Group 4: Refactor tests (large)

Rewrite tests to exercise `Validate()` and assert on `ValidationError`.

- Refactor `validator_test.go` — tests call `Validate()` on a full `Validator`, then use `errors.As` to find specific `ValidationError` values with expected `Check` names
- Remove `Test_BundleValidatorHasAllValidationFns` guard test
- Verify existing regression tests pass (update if error format changes)
- Add tests for `ValidationError` (Error(), Unwrap(), errors.As)

## Task Group 5: Clean up (small)

Final consistency checks.

- Run `make verify`
- Verify no exported `Check*` functions remain in validators package
- Update CLAUDE.md if needed
