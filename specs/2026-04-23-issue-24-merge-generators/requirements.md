# Requirements

## Functional Requirements

- All 11 generators produce identical output to before (regression tests unchanged)
- Resource builder helpers are available in `internal/render/resourceutil/`
- No exported `ResourceGenerator` or `ResourceGenerators` types in any API
- The `generators` package no longer exists
- The `registryv1` package no longer exists
- `Renderer.Render()` produces the same output for the same inputs
- `NewRendererBuilder()` takes no arguments — always uses the standard validator and generators
- Validator lives in `internal/render/validator/`
- `ValidationError` lives in `internal/render/validator/`

## Non-Functional Requirements

- **Simpler architecture** — generators and validator are implementation details of the renderer
- **Test coverage maintained** — same assertions, same edge cases, same error messages
- **No public API changes** — consumers see identical `Renderer`, `Render()`, `NewRendererBuilder()`, etc.
- **Minimal internal API surface** — `BundleValidator` and `Options` are unexported

## Constraints

- Rendering behavior must not change — regression tests pass without golden file changes
- Resource builder helpers must remain testable (hence the `resourceutil` package, not private functions)
- Generator test coverage must be preserved — if individual generator tests are replaced with `Render()` output tests, the same scenarios and assertions must exist

## Dependencies

- Concrete `Renderer` from #18
- All existing internal packages
