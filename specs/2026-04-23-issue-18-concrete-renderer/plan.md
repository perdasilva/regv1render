# Refactor Renderer to Concrete Type with Two-Tier Options

Replace the generic `BundleRenderer` struct + `[]ResourceGenerator` pattern with a concrete `Renderer` type created via `NewRenderer()`. Split options into renderer-level (set once at construction) and render-level (per-call). Simplify the public API by removing internal type aliases.

## Task Group 1: Define new Renderer type and options (medium)

Create the new concrete renderer with two-tier option types.

- Define `Renderer` struct in `internal/render/` that owns a `BundleValidator`, list of resource generators, and renderer-level config (cert provider, unique name generator, deployment config, provided APIs flag)
- Define `RendererOption` functional option type for constructor config
- Define `RenderOption` functional option type for per-render config (target namespaces)
- Implement `NewRenderer(...RendererOption) *Renderer` constructor
- Implement `(*Renderer).Render(rv1, installNamespace, ...RenderOption) ([]client.Object, error)`
- Wire all 10 generators internally — no external generator list

## Task Group 2: Update internal wiring (medium)

Update `registryv1.go` and internal packages.

- Update `registryv1.Renderer` to use the new concrete `Renderer` type
- Remove `BundleRenderer` struct, `ResourceGenerator`, `ResourceGenerators`, `UniqueNameGenerator` types from `internal/render/render.go`
- Remove `Options` struct and `Option` type
- Remove `AdditionalGenerators` workaround
- Move `CertProvisionerFor`, `DefaultUniqueNameGenerator` to unexported or keep internal
- Ensure generators and certproviders still compile with the new internal API

## Task Group 3: Update public API (medium)

Simplify the root package exports.

- Export `Renderer`, `NewRenderer`, `RendererOption`, `RenderOption`
- Keep: `RegistryV1`, `CertificateProvider`, `DeploymentConfig`, `ValidationError`, `BundleSource`, `FromFS`, `FromBundle`, `CertManagerProvider`, `OpenShiftServiceCAProvider`
- `WithCertificateProvider`, `WithDeploymentConfig`, `WithUniqueNameGenerator` become `RendererBuilder` methods
- `WithTargetNamespaces` and `WithProvidedAPIsClusterRoles` become `RenderOption` functions
- Update `Render()` convenience function to use a default `Renderer`
- Update `DefaultRenderer` to be `*Renderer` created by `NewRenderer()`
- Remove: `BundleRenderer`, `BundleValidator`, `ResourceGenerator`, `ResourceGenerators`, `UniqueNameGenerator`, `Options`, `Option`, `CertProvisionerFor`, `DefaultUniqueNameGenerator`, `CertSecretInfo`, `CertificateProvisioner`, `CertificateProvisionerConfig`

## Task Group 4: Update CLI (small)

Update the CLI to use the new API.

- Update `cmd/rv1/render.go` to use `NewRenderer` with config file options, then call `Render` with flag options
- Remove any references to old types

## Task Group 5: Update tests (large)

Update all tests to use the new API.

- Update `internal/render/render_test.go` to use new `Renderer` type
- Update `internal/render/registryv1/registryv1_test.go`
- Update root package tests (`example_test.go`, `providedapis_test.go`)
- Verify regression tests pass (update `generate-manifests.go` if needed)
- Run `make verify`

## Task Group 6: Consolidate files and documentation (small)

- Merge `providedapis.go` into `generators.go` (and tests)
- Fold `deploymentconfig.go` into `render.go`
- Make provided APIs generator flag-driven: always in generator list, checks `opts.ProvidedAPIsClusterRoles`
- Update CLAUDE.md if architecture section needs changes
- Update godoc on new public types
- Run final `make verify`
