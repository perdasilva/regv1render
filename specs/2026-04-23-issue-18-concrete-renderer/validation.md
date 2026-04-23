# Validation

## Acceptance Criteria

1. `NewRenderer()` creates a working renderer
2. `NewRenderer(WithCertificateProvider(...), WithDeploymentConfig(...), WithProvidedAPIsClusterRoles())` applies all options
3. `renderer.Render(rv1, "ns", WithTargetNamespaces("watch"))` works
4. `Render()` convenience function works without options
5. `BundleRenderer`, `ResourceGenerator`, `ResourceGenerators`, `UniqueNameGenerator`, `Options`, `Option` no longer in public API
6. CLI uses `NewRenderer` + `Render` pattern
7. `make verify` passes
8. All regression tests pass without golden file changes
9. `go doc github.com/perdasilva/regv1render` shows simplified API

## Test Scenarios

- `NewRenderer()` + `Render()` — basic rendering works
- `NewRenderer(WithCertificateProvider(CertManagerProvider{}))` + `Render()` — cert provider applied
- `NewRenderer(WithDeploymentConfig(cfg))` + `Render()` — deployment config applied
- `NewRenderer(WithProvidedAPIsClusterRoles())` + `Render()` — provided API roles generated
- `Render()` convenience function — renders with default renderer
- `Render(rv1, ns, WithTargetNamespaces("watch"))` — target namespaces applied
- Invalid bundle — `Render()` returns `ValidationError`

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All regression tests pass unchanged
- `go doc` shows no removed types

## Manual Verification

1. `go doc github.com/perdasilva/regv1render` — verify simplified API
2. `make build && bin/rv1 render --help` — CLI still works
3. Run regression tests — golden files unchanged
