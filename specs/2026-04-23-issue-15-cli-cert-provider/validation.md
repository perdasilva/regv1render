# Validation

## Acceptance Criteria

1. Config file with `certificateProvider: { type: cert-manager }` applies the cert-manager provider
2. Config file with `certificateProvider: { type: openshift-service-ca }` applies the OpenShift provider
3. Config file with `certificateProvider: { type: none }` applies no provider
4. Omitting `certificateProvider` entirely applies no provider (backward compat)
5. Config file with `certificateProvider: { type: invalid }` produces a clear error listing valid types
6. `make verify` passes
7. `make build` produces a working rv1 binary
8. README shows the discriminated union config format

## Test Scenarios

- Parse config with `type: cert-manager` — verify provider option is built
- Parse config with `type: openshift-service-ca` — verify provider option is built
- Parse config with `type: none` — verify no provider option
- Parse config with no `certificateProvider` field — verify no provider (backward compat)
- Parse config with `type: bogus` — verify error message contains "bogus" and lists valid types
- Render a webhook bundle with `type: cert-manager` in config — verify cert-manager objects in output

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All existing tests and regression tests pass unchanged

## Manual Verification

1. Create a config file with `certificateProvider: { type: cert-manager }`
2. Pipe a webhook bundle: `tar -cf - -C test/regression/testdata/bundles/webhook-operator.v0.0.5 . | bin/rv1 render --install-namespace test-ns --config config.yaml`
3. Verify cert-manager Issuer and Certificate objects appear in output
