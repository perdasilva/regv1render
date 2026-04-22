# rv1 CLI Tool

Build the `rv1` CLI that renders registry+v1 bundles to plain Kubernetes manifests. Reads a bundle tar stream from stdin (e.g., piped from `crane export` or `docker export`), applies rendering options from flags and an optional YAML config file, and outputs multi-document YAML to stdout.

## Task Group 1: CLI framework and basic rendering (medium)

Set up the cobra-based CLI with the `render` subcommand and stdin tar reading.

- Add `github.com/spf13/cobra` dependency (already a transitive dep via controller-runtime)
- Replace the placeholder `cmd/rv1/main.go` with a cobra root command and `render` subcommand
- Implement stdin tar reading: read a tar stream, extract it to a temporary `fs.FS`, pass to `regv1render.FromFS()`
- Add `--install-namespace` flag (required)
- Add `--watch-namespace` flag (optional, repeatable)
- Render the bundle using `regv1render.Render()` and output multi-document YAML (`---` separated) to stdout
- Verify basic flow works: `crane export <image> - | rv1 render --install-namespace test-ns`

## Task Group 2: Config file support (medium)

Add YAML config file support for rendering options that are too complex for flags.

- Add `--config` flag that accepts a path to a YAML config file
- Define the config file schema:
  ```yaml
  installNamespace: <string>     # overridden by --install-namespace flag
  watchNamespaces: [<string>]    # overridden by --watch-namespace flag
  providedAPIsClusterRoles: bool
  deploymentConfig:
    nodeSelector: {...}
    tolerations: [...]
    resources: {...}
    env: [...]
    envFrom: [...]
    volumes: [...]
    volumeMounts: [...]
    affinity: {...}
    annotations: {...}
  ```
- Parse the config file and merge with flags (flags take precedence over config file)
- Map config fields to rendering options: `WithTargetNamespaces()`, `WithProvidedAPIsClusterRoles()`, `WithDeploymentConfig()`

## Task Group 3: Error handling and help (small)

Polish the CLI with proper error handling, usage help, and validation.

- Validate that `--install-namespace` is provided (either via flag or config file)
- Print clear error messages for: missing namespace, invalid config file, malformed tar input, render failures
- Add usage examples in the cobra command help text
- Add `--version` flag or subcommand

## Task Group 4: Integration tests (medium)

Write tests that exercise the CLI end-to-end.

- Create test bundles as tar archives in `testdata/`
- Test basic rendering: pipe a tar bundle via stdin, verify YAML output
- Test with config file: verify config file options are applied
- Test flag override: verify flags take precedence over config file
- Test error cases: missing namespace, invalid tar, bad config file
- Verify `make build && bin/rv1 render --help` works

## Task Group 5: Documentation (small)

Update documentation to reflect the CLI capabilities.

- Update README.md with CLI usage examples (crane/docker pipe pattern)
- Update CLAUDE.md if architecture section needs changes
- Run `make verify` — all tests pass
