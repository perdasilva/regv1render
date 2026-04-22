# Port & Refactor Rendering Logic

Port the registry+v1 bundle rendering code from `operator-framework/operator-controller/internal/rukpak/render` into this project. Copy-then-refactor approach: first get the upstream code and tests compiling and passing, then restructure into a thin public API at the repo root with implementation details under `internal/`.

## Task Group 1: Copy upstream code and dependencies (large)

Clone the upstream repo and copy the rendering code verbatim into `internal/`, preserving the original package structure. Get it compiling.

- Clone `operator-framework/operator-controller` at latest main
- Copy `internal/operator-controller/rukpak/render/` into `internal/render/`
- Copy supporting packages into `internal/`: `bundle/` (including `BundleSource` / `FromFS`), `util/`, `config/` (including `DeploymentConfig` / `SubscriptionConfig` types)
- Copy test helpers from `internal/operator-controller/rukpak/util/testing/` into `internal/testutil/`
- Run `go mod tidy` to pull in all required dependencies (k8s, controller-runtime, OLM APIs, operator-registry, cert-manager, jsonschema). If tidy fails, manually add `require`/`replace` directives as needed
- Fix import paths to point to this module's `internal/` packages
- Verify `go build ./...` compiles

## Task Group 2: Port upstream tests (large)

Copy all test files and get them passing.

- Copy test files for render, generators, validators, and cert providers
- Copy test builder helpers (bundlefs, CSV builder)
- Copy any upstream testdata/ directories or fixtures used by tests
- Fix import paths in test files
- Run `go test ./...` and fix any failures
- Verify all upstream test cases pass

## Task Group 3: Port regression tests (medium)

Port the golden-file regression tests from `test/regression/convert/` that verify rendering output byte-for-byte against expected fixtures.

- Copy `test/regression/convert/convert_test.go` and `generate-manifests.go` into a suitable location (e.g., `test/regression/`)
- Copy input bundle fixtures: `testdata/bundles/argocd-operator.v0.6.0/` and `testdata/bundles/webhook-operator.v0.0.5/`
- Copy expected output fixtures: `testdata/expected-manifests/` (107 golden YAML files across 7 test cases)
- Fix import paths to use this module's packages and public API
- Run the regression tests and verify all 7 cases pass:
  - argocd-operator AllNamespaces, SingleNamespace, OwnNamespace
  - webhook-operator all webhook types
  - argocd-operator with DeploymentConfig options, empty Affinity, empty Affinity subtype
- Verify `generate-manifests.go` can regenerate the expected output

## Task Group 4: Define public API surface (medium)

Create the thin public API at the repo root that re-exports the key types and provides a clean entry point.

- Define the public types in root package files: `BundleRenderer`, `BundleValidator`, `ResourceGenerator`, `Options`, `CertificateProvider`
- Export a default `Renderer` (the pre-wired registryv1 renderer with all validators and generators)
- Keep internal implementation details (individual generators, validators, cert provider impls) under `internal/`
- Write godoc comments for all public types, interfaces, and functions
- Verify consumers can import `github.com/perdasilva/regv1render` and render a bundle

## Task Group 5: Clean up and review (small)

Review the ported code for quality, remove dead code, and ensure consistency.

- Remove any upstream code that references operator-controller-specific concerns (e.g., controller reconciler hooks)
- Verify no `internal/` packages import from the root package (dependency direction: root → internal, never reverse)
- Run `make verify` — all linting, vetting, and tests pass
- Run `make build` — rv1 CLI still compiles
- Update CLAUDE.md if the architecture section needs changes
