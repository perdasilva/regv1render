# Requirements

## Functional Requirements

- CLI config file supports a `certificateProvider` field with discriminated union structure
- `type: cert-manager` maps to `CertManagerProvider`
- `type: openshift-service-ca` maps to `OpenShiftServiceCAProvider`
- `type: none` or omitted `certificateProvider` means no provider (default)
- Unknown `type` values produce a clear error message
- The struct is extensible for future provider-specific config (e.g., `secret` provider in #19)

## Non-Functional Requirements

- **Backward compatible** — omitting `certificateProvider` behaves identically to current behavior
- **Extensible** — adding new provider types (e.g., `secret`) requires only adding fields to the struct, not changing the parsing logic
- **Clear errors** — invalid `type` values list the valid options in the error message

## Constraints

- No library changes — this is purely CLI config plumbing
- Use `sigs.k8s.io/yaml` for parsing (already a dependency)
- The `CertificateProviderConfig` struct must use JSON tags compatible with YAML parsing
- Do not add provider-specific config fields yet — that's #19

## Dependencies

- `rv1` public API: `WithCertificateProvider()`, `CertManagerProvider`, `OpenShiftServiceCAProvider`
