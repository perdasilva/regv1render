# Requirements

## Functional Requirements

- README contains a "Namespace Modes" section explaining AllNamespaces, OwnNamespace, SingleNamespace, and MultiNamespace
- Each mode has CLI flag examples
- `rv1 render --help` describes namespace behavior and default
- A demo script (`hack/demo.sh`) exists that exercises each namespace mode
- Instructions for recording and converting the demo to GIF exist in `hack/README.md`
- README has a placeholder for the demo GIF

## Non-Functional Requirements

- Documentation is clear enough for a user unfamiliar with OLM install modes
- CLI examples are copy-pasteable
- Demo script is runnable with a local bundle tar

## Constraints

- No library or CLI code changes — documentation only (except help text update)
- Do not change rendering behavior
- Demo script should work with the existing test bundle fixtures

## Dependencies

- rv1 CLI from epic #4
- Test bundle fixtures in `test/regression/testdata/bundles/`
