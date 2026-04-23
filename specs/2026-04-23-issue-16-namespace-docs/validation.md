# Validation

## Acceptance Criteria

1. README contains a "Namespace Modes" section with 4 modes explained
2. Each mode has CLI flag examples
3. `rv1 render --help` describes namespace behavior
4. `assets/demo.gif` exists
5. README embeds the demo GIF
7. `make verify` passes
8. `make build` still works

## Test Scenarios

- Read the "Namespace Modes" section — verify all 4 modes are covered
- Run `bin/rv1 render --help` — verify namespace info appears
- Verify `assets/demo.gif` renders correctly in the README

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All existing tests pass unchanged

## Manual Verification

1. Read README end-to-end — verify namespace section flows naturally
2. Run `bin/rv1 render --help` — verify help text is clear
3. Verify `assets/demo.gif` is visible when viewing README on GitHub
