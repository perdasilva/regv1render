# Documentation & Examples

Provide comprehensive documentation and usage examples for the regv1render library and rv1 CLI. Add testable examples that appear in godoc, a CONTRIBUTING.md for contributors, and enhance the README with upstream relationship context.

## Task Group 1: Testable examples for godoc (medium)

Create `Example*` functions in `_test.go` files at the repo root that demonstrate common use cases. These are compiled and verified by `go test` and rendered in godoc.

- `ExampleRender` — basic rendering of a bundle with install namespace
- `ExampleRender_withTargetNamespaces` — rendering with specific watch namespaces
- `ExampleRender_withProvidedAPIsClusterRoles` — rendering with OLMv0 provided API roles enabled
- `ExampleFromFS` — loading a bundle from an `fs.FS`
- `ExampleFromBundle` — creating a BundleSource from an already-parsed RegistryV1 struct
- `ExampleDefaultRenderer` — using the DefaultRenderer directly for more control
- Each example should use realistic but minimal bundle data
- Verify all examples pass: `go test -run Example ./`

## Task Group 2: Godoc audit (small)

Review and improve godoc comments across the public API.

- Check all exported types, functions, and variables in `render.go`, `source.go`, `certproviders.go`, `doc.go`
- Ensure doc comments are complete, accurate, and follow Go conventions (start with the name of the thing being documented)
- Add package-level documentation to `doc.go` describing the library's purpose, relationship to upstream, and basic usage
- Run `go doc github.com/perdasilva/regv1render` and verify the output is clean

## Task Group 3: README enhancements (small)

Enhance the README with upstream relationship context and polish.

- Add a section explaining the relationship to `operator-framework/operator-controller` and `operator-framework/operator-lifecycle-manager`
- Add a section on the OLMv0 compatibility option (`WithProvidedAPIsClusterRoles`)
- Review existing sections for accuracy after all epics
- Ensure the README flows logically: overview → install → library usage → CLI usage → OLMv0 compat → development → license

## Task Group 4: CONTRIBUTING.md (small)

Create a contributor guide covering the development workflow.

- Prerequisites (Go 1.25+, bingo manages dev tools)
- Clone and build instructions
- Running tests (`make verify`)
- Branch naming and commit message conventions (reference specs/conventions.md)
- PR process (Summary + Test Plan format)
- Brief reference to the SDD workflow for maintainers

## Task Group 5: Final review (small)

Verify all documentation is consistent and accurate.

- Run `make verify` — all tests (including examples) pass
- Run `go doc github.com/perdasilva/regv1render` — verify clean output with examples
- Read through README and CONTRIBUTING end-to-end for coherence
- Update CLAUDE.md if needed
