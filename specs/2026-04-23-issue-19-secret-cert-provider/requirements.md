# Requirements

## Functional Requirements

- `SecretCertificateProvider` implements the `CertificateProvider` interface
- `InjectCABundle` is a no-op (returns nil)
- `AdditionalObjects` returns a `kubernetes.io/tls` Secret with `tls.crt` and `tls.key` data fields
- If cert/key data is provided, the Secret contains that data; if not, the fields are empty
- `GetCertSecretInfo` returns the secret name and key names (`tls.crt`, `tls.key`)
- CLI config supports `certificateProvider: { type: secret }` with optional nested `secret: { cert: ..., key: ... }`
- The existing generator mounts the secret into webhook deployments (no changes needed)

## Non-Functional Requirements

- **Consistent with existing providers** — follows the same pattern as CertManager and OpenShiftServiceCA providers
- **Backward compatible** — existing config files continue to work unchanged
- **Well-tested** — unit tests for the provider, CLI config tests, regression test with webhook bundle

## Constraints

- Do not modify the `CertificateProvider` interface
- Do not modify existing providers (CertManager, OpenShiftServiceCA)
- Keep the provider stateless except for the cert/key data it holds

## Dependencies

- `CertificateProvider` interface from `internal/render/certprovider.go`
- `CertProviderResourceGenerator` (already handles Secret mounting)
- Discriminated union config from #15
