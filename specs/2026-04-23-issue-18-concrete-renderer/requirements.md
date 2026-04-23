# Requirements

## Functional Requirements

- `NewRenderer(...RendererOption)` creates a `*Renderer` with static configuration
- `(*Renderer).Render(rv1, installNamespace, ...RenderOption)` validates and renders a bundle
- `RendererOption` covers: cert provider, deployment config, provided APIs ClusterRoles, unique name generator
- `RenderOption` covers: target/watch namespaces
- `Render()` convenience function uses a default `Renderer` (no options)
- `DefaultRenderer` is a `*Renderer` created by `NewRenderer()`
- All 10 generators run internally — not externally wired
- Validation uses the concrete `validator.BundleValidator` from #17
- Rendering output is identical to the previous implementation for the same inputs

## Non-Functional Requirements

- **Simpler API** — public surface reduced from ~20 type aliases to ~12 core exports
- **Clear option tiers** — renderer config (set once) vs render config (per-call) are distinct types
- **CLI alignment** — config file maps to `RendererOption`, flags map to `RenderOption`
- **Backward-compatible output** — same bundles produce the same manifests

## Constraints

- Rendering behavior must not change — regression tests pass without golden file changes
- Keep `CertificateProvider` as a public interface (needed by consumers who implement custom providers)
- Keep `DeploymentConfig` public (used in CLI config files)
- Generators stay in `internal/render/registryv1/generators/` — don't move them

## Dependencies

- Concrete validator from #17
- All existing internal packages (generators, certproviders, bundle, etc.)
