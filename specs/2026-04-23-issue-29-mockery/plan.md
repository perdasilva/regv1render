# Replace Custom Fakes with testify/mock and Restructure Packages

Replace all hand-rolled fake/stub types with generated mocks using testify/mock and the mockery code generator. Restructure internal packages for cleaner dependency graph.

## Task Group 1: Add mockery tooling (small)

- Add mockery to `.bingo/` via `bingo get github.com/vektra/mockery/v2`
- Add a `.mockery.yaml` config file at the repo root
- Add a `generate` target to the Makefile
- Run mockery and verify the generated mock compiles
- Verify `make verify` passes

## Task Group 2: Move CertificateProvider interface to certproviders (medium)

Move the interface and related types to live alongside their implementations, breaking the import cycle that prevented the mock from living with the interface.

- Move `CertificateProvider`, `CertSecretInfo`, `CertificateProvisionerConfig`, `CertificateProvisioner` from `internal/render/` to `internal/render/certproviders/`
- Extract shared utilities (`DeepHashObject`, `ObjectNameForBaseAndSuffix`, `ToUnstructured`, `MergeMaps`) to `internal/renderutil/`
- Delete `internal/render/util.go` (was just wrappers)
- Update all callers to use `renderutil.` and `certproviders.` prefixes
- Merge `certProvisionerFor` into `renderer.go`, delete `certprovider.go`
- Generate mock into `internal/render/certproviders/` alongside the interface

## Task Group 3: Replace fakes with generated mock (medium)

- Replace all `FakeCertProvider` usage in test files with `MockCertificateProvider` + EXPECT patterns
- Replace `rendererTestCertProvider` with the generated mock
- Delete `internal/render/fake.go`
- Remove `FakeCertProvider` from testutil
- Move certprovider provisioner tests to `renderer_internal_test.go`

## Task Group 4: Restructure packages (small)

- Merge `internal/render/resourceutil/` into `internal/renderutil/`
- Flatten `internal/util/testutil/` to `internal/testutil/`
- Unexport `CertProvisionerFor` → `certProvisionerFor`
- Rename `registryv1.go` → `bundle.go`
- Update CLAUDE.md architecture section
