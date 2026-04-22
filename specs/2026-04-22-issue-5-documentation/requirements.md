# Requirements

## Functional Requirements

- Testable examples exist for the core public API: `Render`, `FromFS`, `FromBundle`, `WithTargetNamespaces`, `WithProvidedAPIsClusterRoles`
- All examples compile and pass as part of `go test`
- All examples appear in `go doc` output
- `CONTRIBUTING.md` exists with development workflow, conventions, and PR process
- README documents the upstream relationship to operator-framework
- README documents the OLMv0 compatibility option

## Non-Functional Requirements

- **Godoc quality** — all public types, interfaces, and functions have doc comments that follow Go conventions
- **Example quality** — examples are realistic, minimal, and self-documenting
- **README readability** — logical flow, no broken links, accurate commands

## Constraints

- Do not add new library features — this epic is documentation only
- Examples should use testable `Example*` functions, not standalone programs
- Do not duplicate content between README, CONTRIBUTING, and godoc — cross-reference where appropriate

## Dependencies

- All prior epics (#1-#4) must be complete (library, OLMv0 compat, CLI)
