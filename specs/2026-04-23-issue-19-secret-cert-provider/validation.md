# Validation

## Acceptance Criteria

1. `SecretCertificateProvider{}` implements `CertificateProvider` interface
2. `AdditionalObjects` produces a `kubernetes.io/tls` Secret with correct keys
3. With cert/key data provided, the Secret contains that data
4. Without cert/key data, the Secret has empty `tls.crt`/`tls.key`
5. `InjectCABundle` returns nil (no-op)
6. CLI config `certificateProvider: { type: secret }` works
7. CLI config `certificateProvider: { type: secret, secret: { cert: ..., key: ... } }` works
8. `make verify` passes
9. Regression test with webhook bundle + secret provider passes

## Test Scenarios

- Create `SecretCertificateProvider{}` with no data, call `AdditionalObjects` — verify Secret with empty data
- Create `SecretCertificateProvider{Cert: ..., Key: ...}` with data, call `AdditionalObjects` — verify Secret with provided data
- Verify Secret type is `kubernetes.io/tls`
- Verify `InjectCABundle` returns nil
- Verify `GetCertSecretInfo` returns correct secret name and key names
- Parse CLI config with `type: secret` — verify provider is created
- Parse CLI config with `type: secret` and cert/key data — verify data flows to provider
- Render webhook bundle with secret provider — verify Secret appears in output

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All existing tests and regression tests pass unchanged

## Manual Verification

1. Create config with `certificateProvider: { type: secret }`
2. Render webhook bundle: `tar -cf - -C test/regression/testdata/bundles/webhook-operator.v0.0.5 . | bin/rv1 render --install-namespace test-ns --config config.yaml`
3. Verify a `kubernetes.io/tls` Secret appears in the output
