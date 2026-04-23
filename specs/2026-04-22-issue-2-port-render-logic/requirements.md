# Requirements

## Functional Requirements

- The library can render a registry+v1 bundle into a list of plain Kubernetes manifests
- All upstream validators are ported and functional (uniqueness, DNS naming, install mode compat, webhook rules, etc.)
- All upstream resource generators are ported and functional (deployments, RBAC, CRDs, webhooks, services, cert resources)
- CertificateProvider interface is available with cert-manager and OpenShift service-ca implementations
- Bundle loading from `fs.FS` is supported via a ported `BundleSource` interface
- DeploymentConfig / SubscriptionConfig support is ported
- All upstream unit tests pass against the ported code
- All upstream regression tests pass — golden-file tests verify rendered output byte-for-byte against 107 expected YAML fixtures across 7 test cases
- The public API is importable at `github.com/perdasilva/rv1`

## Non-Functional Requirements

- **Upstream fidelity** — rendering output must be byte-for-byte identical to upstream for the same inputs
- **Clean API** — public surface at root should be minimal; implementation in `internal/`
- **Godoc quality** — all public types, interfaces, and functions have doc comments
- **Dependency direction** — root package may import `internal/`; `internal/` must never import root

## Constraints

- Do not modify upstream rendering logic — this epic is a faithful port, not an improvement
- Do not remove or alter any upstream test cases
- Port from latest main of `operator-framework/operator-controller`
- Preserve upstream package structure inside `internal/` where it aids code navigation
- The rv1 CLI should still compile but does not need to use the rendering API yet

## Dependencies

- `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/apiextensions-apiserver`, `k8s.io/cli-runtime`, `k8s.io/utils`
- `sigs.k8s.io/controller-runtime` (for `client.Object`)
- `github.com/operator-framework/api` (OLM API types — CSV, SubscriptionConfig)
- `github.com/operator-framework/operator-registry` (bundle annotations, property types)
- `github.com/cert-manager/cert-manager` (cert-manager CRD types)
- `github.com/santhosh-tekuri/jsonschema/v6` (config validation)
