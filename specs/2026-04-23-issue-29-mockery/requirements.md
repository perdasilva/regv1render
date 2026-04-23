# Requirements

## Functional Requirements

- mockery is pinned via `.bingo/` and invocable via `make generate`
- A generated mock exists for the `CertificateProvider` interface in `internal/render/certproviders/`
- All test files use the generated mock instead of hand-rolled fakes
- `internal/render/fake.go` is deleted
- `FakeCertProvider` is removed from all packages
- `CertificateProvider` interface and related types live in `certproviders` package
- Shared utilities live in `internal/renderutil/` (one package, not two)
- Test utilities live in `internal/testutil/` (not `internal/util/testutil/`)
- All existing test assertions and scenarios are preserved

## Non-Functional Requirements

- **Clean dependency graph** — no cycles, clear direction: `render` → `certproviders` → `renderutil`
- **Consistency** — one way to mock interfaces, one utility package
- **Minimal exported surface** — `CertProvisionerFor` unexported, `options` unexported

## Constraints

- No public API changes
- Preserve `ToUnstructuredT` and `FakeBundleSource` in testutil
- Generated mocks committed to version control

## Dependencies

None
