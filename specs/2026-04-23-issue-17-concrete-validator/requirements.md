# Requirements

## Functional Requirements

- A concrete `Validator` struct replaces `BundleValidator []func(*RegistryV1) []error`
- `Validator.Validate(*RegistryV1)` runs all 13 checks and returns a joined error
- Each validation failure is wrapped in a `ValidationError` with a `Check` field identifying which check failed
- `ValidationError` implements `Error()` and `Unwrap()` for standard Go error handling
- Consumers can use `errors.As(err, &ve)` to extract `ValidationError` and inspect `ve.Check`
- All 13 existing checks produce identical validation behavior (same inputs fail/pass)

## Non-Functional Requirements

- **Simpler API** — consumers no longer see `BundleValidator`, `ResourceGenerator`, or the composition pattern
- **Structured errors** — programmatic error handling via `ValidationError.Check` instead of string parsing
- **Backward compatible rendering** — `Render()` and `DefaultRenderer` continue to work identically
- **Testable** — tests exercise the real `Validate()` path, not individual functions in isolation

## Constraints

- Keep the 13 check functions in `internal/render/registryv1/validators/` — don't move them to another package
- Validation behavior must not change — same bundles pass/fail the same way
- Regression tests must continue to pass without golden file changes
- `ValidationError` type must be exported from the root package for consumer use

## Dependencies

- Existing rendering pipeline from epic #2
- `internal/render/registryv1/validators/validator.go` — 13 check functions
- `internal/render/registryv1/validators/validator_test.go` — 1254 lines of tests
