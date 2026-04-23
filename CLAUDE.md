# rv1

A standalone Go library for rendering OLM registry+v1 bundles to plain Kubernetes manifests. Extracted from `operator-framework/operator-controller/internal/rukpak/render` and compatible with `operator-framework/operator-lifecycle-manager` rendering behavior.

## Architecture

```
*.go           Public API at repo root — consumers import github.com/perdasilva/rv1
testdata/      Test fixtures (ignored by go build)
internal/      Non-public implementation details (not importable by consumers)
  bundle/      RegistryV1 bundle type, annotations, and source loading (fs.FS)
  render/      Core rendering engine, generators, cert providers, resource builders
    validator/     Bundle validation logic
    resourceutil/  Kubernetes resource builder helpers (Deployment, Role, Service, etc.)
    certproviders/ Certificate provider implementations (cert-manager, service-ca, secret)
  util/testutil/ Test helpers (bundlefs builder, CSV builder)
cmd/rv1/       Showcase CLI tool for rendering bundles from the command line
test/          Regression tests with golden-file fixtures
.bingo/        Pinned dev tool versions (managed by bingo)
specs/         SDD governing specs (mission, tech stack, conventions)
```

## Design Principles

- Upstream fidelity — output must match operator-controller and operator-lifecycle-manager behavior
- Minimal dependencies — only import what's strictly necessary
- Composable API — expose building blocks, not a monolithic render function
- Test-driven — upstream test cases ported alongside logic
- Makefile-driven — all operations via Make targets

## Build Commands

```
make build    Build the rv1 CLI binary
make test     Run all unit tests
make lint     Run golangci-lint
make fmt      Run gofmt and goimports
make vet      Run go vet
make tidy     Run go mod tidy
make verify   Full quality gate (fmt + vet + lint + test)
make clean    Remove build artifacts
```

## Epic-Based Workflow

Work is organized into epics tracked as GitHub issues with the `epic` label. Use these slash commands to drive the workflow:

| Command | Purpose |
|---|---|
| `/sdd-plan-next-phase` | Find the next epic issue, create a branch, write a detailed spec |
| `/sdd-implement` | Implement the current epic following its spec |
| `/sdd-review` | Review all branch changes for quality and spec compliance |
| `/sdd-ship` | Verify, commit, and publish — creates PR with `Closes #N` |
| `/sdd-ideate` | Brainstorm and create new epics |

## Conventions

- **Commits:** Conventional commits (`feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`)
- **Branches:** `type/short-description` (e.g., `feat/registry-v1-parser`)
- **PRs:** Title matches commit subject; body has Summary + Test Plan sections
