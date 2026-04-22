# regv1render

A standalone Go library for rendering OLM registry+v1 bundles to plain Kubernetes manifests.

Extracted from [`operator-framework/operator-controller/internal/rukpak/render`](https://github.com/operator-framework/operator-controller) and compatible with [`operator-framework/operator-lifecycle-manager`](https://github.com/operator-framework/operator-lifecycle-manager) rendering behavior.

## Install

```bash
go get github.com/perdasilva/regv1render
```

### CLI

```bash
go install github.com/perdasilva/regv1render/cmd/rv1@latest
```

## Development

Requires Go 1.24+. Dev tools are managed via [bingo](https://github.com/bwplotka/bingo) and built automatically on first use.

```bash
make build     # build the rv1 CLI binary
make test      # run all unit tests
make lint      # run golangci-lint
make fmt       # run gofmt and goimports
make vet       # run go vet
make verify    # full quality gate (fmt + vet + lint + test)
make clean     # remove build artifacts
```

## License

Apache-2.0
