# Secret Certificate Provider

Add a "secret" CertificateProvider that generates a `kubernetes.io/tls` Secret with optional user-provided cert/key data. Extend the CLI config's `certificateProvider` discriminated union with the `secret` type and nested config.

## Task Group 1: Implement SecretCertProvider (medium)

Create the provider in `internal/render/certproviders/`.

- Create `SecretCertProvider` struct with `Cert []byte` and `Key []byte` fields
- `InjectCABundle` — no-op (return nil)
- `AdditionalObjects` — return a `kubernetes.io/tls` Secret with `tls.crt`/`tls.key` data (empty if not provided)
- `GetCertSecretInfo` — return the secret name and key names (`tls.crt`, `tls.key`)
- Export `SecretCertProvider` type from `render.go` (direct alias, no indirection)
- Remove `certproviders.go` — all provider type exports consolidated in `render.go`

## Task Group 2: Update CLI config (small)

Extend the discriminated union to support the `secret` type with nested config.

- Add `Secret *secretProviderConfig` field to `certificateProviderConfig` struct
- Define `secretProviderConfig` with `Cert string` and `Key string` fields
- Add `"secret"` to `validCertProviderTypes`
- Wire `secret` type in `buildRenderer` to create `SecretCertProvider` with cert/key data
- Add `certProviderSecret` constant

## Task Group 3: Add tests (medium)

- Unit tests for `SecretCertProvider`: empty data, provided data, secret type/name/keys
- CLI config tests: parse secret config with cert/key, parse secret config without cert/key
- Regression test: render webhook bundle with secret provider, verify Secret object in output (11 golden files)

## Task Group 4: Documentation (small)

- Update README with detailed secret provider config example (with and without cert/key data)
- Update `rv1 render --help` to mention secret provider
- Update godoc on `SecretCertProvider`
