# Tech Stack

## Language & Runtime

- **Language:** Go 1.25
- **Module path:** `github.com/perdasilva/regv1render`

## Project Structure

```
regv1render/
├── render.go            # public API entry points
├── types.go             # public types and interfaces
├── render_test.go       # tests
├── testdata/            # test fixtures (ignored by go build)
├── internal/            # non-public implementation details (not importable by consumers)
├── cmd/
│   └── rv1/             # showcase CLI tool
│       └── main.go
├── specs/               # SDD governing specs and per-epic specs
│   ├── mission.md
│   ├── tech-stack.md
│   ├── conventions.md
│   └── YYYY-MM-DD-issue-N-*/   # epic specs created by /sdd-plan-next-phase
│       ├── plan.md
│       ├── requirements.md
│       └── validation.md
├── .claude/
│   └── commands/        # SDD workflow commands
├── .github/
│   └── workflows/       # GitHub Actions CI
│       └── ci.yml
├── .bingo/              # pinned dev tool versions (managed by bingo)
├── .golangci.yml        # golangci-lint configuration
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md
└── .gitignore
```

## Dependencies

### Core

| Dependency | Purpose |
|---|---|
| `k8s.io/api` | Kubernetes API types (core, apps, rbac, etc.) |
| `k8s.io/apimachinery` | API machinery (unstructured, scheme, serialization) |
| `sigs.k8s.io/controller-runtime` | Controller runtime client types |
| `sigs.k8s.io/yaml` | YAML serialization |

### Dev (managed via [bingo](https://github.com/bwplotka/bingo))

| Tool | Purpose |
|---|---|
| `golangci-lint` | Linting |
| `goimports` | Import formatting |

Tool versions are pinned in `.bingo/` and auto-built on first use via `make`. Add new tools with `bingo get <module>`.

## Build Commands

| Command | Description |
|---|---|
| `make build` | Build the rv1 CLI binary |
| `make test` | Run all unit tests |
| `make lint` | Run golangci-lint |
| `make fmt` | Run gofmt and goimports |
| `make vet` | Run go vet |
| `make tidy` | Run go mod tidy |
| `make verify` | Run fmt + vet + lint + test (full quality gate) |
| `make clean` | Remove build artifacts |

## CI/CD

- **GitHub Actions** workflow on push and pull request to main
- Steps: checkout, setup Go 1.25, `make verify`, `make build`
- No container build at this time
