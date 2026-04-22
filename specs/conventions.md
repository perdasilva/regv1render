# Conventions

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>: <short description>

<optional body explaining why>
```

### Types

| Type | When to use |
|---|---|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `refactor` | Code restructuring without behavior change |
| `test` | Adding or updating tests |
| `docs` | Documentation changes |
| `chore` | Build, CI, tooling, dependencies |

### Examples

```
feat: add registry+v1 bundle parser

fix: handle missing CRD annotations in bundle metadata

chore: configure golangci-lint with project rules

refactor: extract manifest sorting into standalone function

test: add edge cases for empty bundle rendering
```

### Rules

- Subject line: imperative mood, lowercase, no period, under 72 characters
- Body (optional): explain *why*, not *what*
- No co-author trailers required

## Pull Requests

### Title

Match the primary commit's subject line (conventional commit format).

### Description

```markdown
## Summary
- <1-3 bullet points describing what changed and why>

## Test Plan
- [ ] <How to verify the changes>
```

### Example

```markdown
## Summary
- Port registry+v1 bundle rendering logic from operator-controller
- Refactor rendering logic into public API at repo root with internal/ for private code

## Test Plan
- [ ] `make check` passes
- [ ] Upstream test cases ported and passing
- [ ] Public API reviewed for minimal surface area
```

## Branch Naming

Format: `<type>/<short-description>`

| Example | When |
|---|---|
| `feat/registry-v1-parser` | New feature |
| `fix/yaml-parse-error` | Bug fix |
| `chore/ci-setup` | Build/tooling work |
| `refactor/split-render-steps` | Refactoring |
| `docs/api-examples` | Documentation |

Rules:
- Lowercase, hyphens for spaces
- Keep under 50 characters
- Use the same type prefix as the commit
