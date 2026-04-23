# CLI Certificate Provider Config

Add a `certificateProvider` discriminated union field to the rv1 CLI config file. Users select a provider via the `type` discriminator (`cert-manager`, `openshift-service-ca`, `none`), and future providers (#19) can add nested config without changing the schema. This is purely CLI config plumbing — no library changes.

## Task Group 1: Config struct and parsing (small)

Add the discriminated union struct and wire it into the config parser.

- Add `CertificateProviderConfig` struct with `Type` field to `cmd/rv1/render.go`
- Add `CertificateProvider *CertificateProviderConfig` field to `renderConfig`
- Map `type` values in `buildRenderOptions()`:
  - `cert-manager` → `regv1render.WithCertificateProvider(regv1render.CertManagerProvider{})`
  - `openshift-service-ca` → `regv1render.WithCertificateProvider(regv1render.OpenShiftServiceCAProvider{})`
  - `none` or nil → no provider (default, current behavior)
- Validate `type` — reject unknown values with a clear error message

## Task Group 2: Tests (small)

Add integration tests for each provider type and error cases.

- Test `cert-manager` type: verify config parses and `WithCertificateProvider` is applied
- Test `openshift-service-ca` type: same
- Test `none` type: verify no provider is applied (same as omitted)
- Test omitted `certificateProvider`: verify backward compat (no provider)
- Test invalid type value: verify clear error message

## Task Group 3: Documentation (small)

Update docs to show the new config format.

- Update README config file example to show discriminated union format
- Update `rv1 render --help` long description with cert provider config info
- Run `make verify` — all tests pass
