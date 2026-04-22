# Mission

A standalone Go library for rendering OLM registry+v1 bundles to plain Kubernetes manifests, extracted from the `operator-framework/operator-controller` internal rendering logic and compatible with the original `operator-framework/operator-lifecycle-manager` rendering behavior.

## Goals

- **Faithful port of registry+v1 rendering logic** from `operator-framework/operator-controller/internal/rukpak/render`
- **OLMv0 rendering compatibility** — respects rendering behavior from `operator-framework/operator-lifecycle-manager`
- **Clean, minimal public API surface** — composable building blocks, not just a monolithic render function
- **Zero dependency on operator-controller internals** — fully standalone library
- **Well-tested** — upstream test cases ported alongside the logic, with equivalent or better coverage

## Non-Goals

- **No bundle fetching/downloading** — input is already-loaded bundle content
- **No support for other bundle formats** (e.g., plain, helm) — registry+v1 only
- **No cluster interaction** — pure offline rendering
- The `rv1` CLI is a showcase tool, not the primary deliverable — other CLIs/tools should provide a more comprehensive experience

## Design Principles

- **Upstream fidelity** — rendering output must match operator-controller and operator-lifecycle-manager behavior
- **Minimal dependencies** — only import what's strictly necessary
- **Composable API** — expose building blocks, not just a monolithic render function
- **Test-driven** — upstream test cases should be ported alongside the logic
- **Makefile-driven** — all build, test, and CI operations via Makefile targets
- **Go best practices** — idiomatic Go, standard project layout, proper error handling

## Development Practices

- `go vet` and `golangci-lint` must pass on all code
- `gofmt`/`goimports` for consistent formatting
- Unit tests with `go test` for all packages
- Code coverage tracking
- GitHub Actions CI on all pushes and PRs
