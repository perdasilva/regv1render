# Requirements

## Functional Requirements

- `rv1 render` reads a registry+v1 bundle as a tar stream from stdin
- `rv1 render --install-namespace <ns>` sets the install namespace (required)
- `rv1 render --watch-namespace <ns>` sets target/watch namespaces (optional, repeatable)
- `rv1 render --config <path>` loads rendering options from a YAML config file
- Flags take precedence over config file values for `installNamespace` and `watchNamespaces`
- Config file supports: `installNamespace`, `watchNamespaces`, `providedAPIsClusterRoles`, `deploymentConfig`
- Output is multi-document YAML (`---` separated) written to stdout
- Errors are written to stderr with clear messages
- `rv1 render --help` prints usage with examples
- `rv1 --version` prints the version

## Non-Functional Requirements

- **Simple UX** — the CLI is a thin wrapper around the library; complexity belongs in the library, not the CLI
- **Composable** — stdin input allows piping from any image tool (`crane export`, `docker export`, `oc image extract`, etc.)
- **Minimal dependencies** — cobra is the only new dependency (already a transitive dep)

## Constraints

- Stdin only — no built-in image pulling; users pipe the content
- No cluster interaction — the CLI is purely offline
- The config file schema should use the same field names as the Go API where possible
- Do not add rendering logic to the CLI — it should only parse input, build options, and call the library

## Dependencies

- `github.com/spf13/cobra` (CLI framework, already a transitive dep)
- `sigs.k8s.io/yaml` (config file parsing, already a dep)
- `rv1` public API (FromFS, Render, With* options)
