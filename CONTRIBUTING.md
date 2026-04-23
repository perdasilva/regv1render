# Contributing to rv1

## Prerequisites

- Go 1.25+
- Dev tools (golangci-lint, goimports) are managed by [bingo](https://github.com/bwplotka/bingo) and built automatically on first `make` invocation — no manual install needed

## Getting Started

```bash
git clone https://github.com/perdasilva/rv1.git
cd rv1
make verify   # run the full quality gate (fmt + vet + lint + test)
make build    # build the rv1 CLI binary
```

## Making Changes

1. Create a branch following the naming convention: `<type>/<short-description>`
   - Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`
   - Example: `feat/new-option`, `fix/crd-parsing`, `docs/api-examples`

2. Make your changes and ensure they pass:
   ```bash
   make verify
   ```

3. Commit using [conventional commits](https://www.conventionalcommits.org/):
   ```
   feat: add support for custom label injection
   
   Optional body explaining why, not what.
   ```

## Pull Requests

- Title should match your commit subject (conventional commit format)
- Include a description with:

```markdown
## Summary
- <what changed and why>

## Test Plan
- [ ] <how to verify>
```

## Build Commands

| Command | Description |
|---|---|
| `make build` | Build the rv1 CLI binary |
| `make test` | Run all unit tests |
| `make lint` | Run golangci-lint |
| `make fmt` | Run gofmt and goimports |
| `make vet` | Run go vet |
| `make verify` | Full quality gate (fmt + vet + lint + test) |
| `make clean` | Remove build artifacts |

## Adding Dev Tools

Dev tool versions are pinned in `.bingo/`. To add a new tool:

```bash
go install github.com/bwplotka/bingo@latest
bingo get <tool-module-path>
```

## Project Structure

```
*.go           Public API (consumers import github.com/perdasilva/rv1)
internal/      Private implementation (render engine, bundle types, utilities)
cmd/rv1/       CLI tool
test/          Regression tests with golden-file fixtures
specs/         SDD governing specs and per-epic specs
```
