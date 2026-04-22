# Validation

## Acceptance Criteria

1. `rv1 render --install-namespace test-ns < bundle.tar` reads stdin and outputs multi-doc YAML to stdout
2. `rv1 render --config config.yaml` reads config from file and applies all options
3. Flags override config file values (e.g., `--install-namespace` overrides `installNamespace` in config)
4. `rv1 render --help` prints usage with examples
5. `rv1 --version` prints version information
6. Missing `--install-namespace` (and not in config) produces a clear error
7. Invalid tar input produces a clear error
8. `make verify` passes
9. `make build` produces a working `bin/rv1`

## Test Scenarios

- Pipe a valid bundle tar to `rv1 render --install-namespace test-ns` — verify YAML output contains expected resources
- Pipe a valid bundle tar with `--watch-namespace watch-ns` — verify target namespace is applied
- Use a config file with `providedAPIsClusterRoles: true` — verify ClusterRoles appear in output
- Use a config file with `deploymentConfig` — verify deployment customizations applied
- Pass `--install-namespace` flag AND `installNamespace` in config — verify flag wins
- Pass `--watch-namespace` flag AND `watchNamespaces` in config — verify flag wins
- Omit `--install-namespace` and config — verify error message
- Pipe invalid data to stdin — verify error message
- Pass non-existent config file — verify error message
- Run `rv1 render --help` — verify usage text and examples
- Run `rv1 --version` — verify version output

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All existing unit and regression tests pass unchanged
- CLI integration tests pass

## Manual Verification

1. Build the CLI: `make build`
2. Render a real bundle: `crane export quay.io/operatorhubio/argocd-operator:v0.6.0 - | bin/rv1 render --install-namespace argocd-system`
3. Verify output is valid multi-document YAML
4. Verify `bin/rv1 render --help` shows clear usage
5. Verify `bin/rv1 --version` prints version
