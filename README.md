# regv1render

A standalone Go library for rendering OLM registry+v1 bundles to plain Kubernetes manifests.

Extracted from [`operator-framework/operator-controller/internal/rukpak/render`](https://github.com/operator-framework/operator-controller) and compatible with [`operator-framework/operator-lifecycle-manager`](https://github.com/operator-framework/operator-lifecycle-manager) rendering behavior.

## Install

```bash
go get github.com/perdasilva/regv1render
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/perdasilva/regv1render"
)

func main() {
	// Load a bundle from a directory on disk
	bundleSource := regv1render.FromFS(os.DirFS("path/to/bundle"))
	rv1, err := bundleSource.GetBundle()
	if err != nil {
		panic(err)
	}

	// Render the bundle to plain Kubernetes manifests
	objects, err := regv1render.Render(rv1, "my-namespace",
		regv1render.WithTargetNamespaces("watch-ns"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Rendered %d objects\n", len(objects))
}
```

### CLI

The `rv1` CLI renders bundles from the command line. It reads a bundle tar stream from stdin:

```bash
go install github.com/perdasilva/regv1render/cmd/rv1@latest

# Render from a container image using docker
docker export $(docker create quay.io/my/bundle:v1 /bin/true) | rv1 render --install-namespace my-ns

# Or using crane
crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns

# Render with watch namespaces
crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns --watch-namespace ns1

# Render with a config file
crane export quay.io/my/bundle:v1 - | rv1 render --config render.yaml
```

Config file format (`render.yaml`):

```yaml
installNamespace: my-ns
watchNamespaces:
  - ns1
providedAPIsClusterRoles: true
deploymentConfig:
  nodeSelector:
    kubernetes.io/os: linux
```

## Development

Requires Go 1.25+. Dev tools are managed via [bingo](https://github.com/bwplotka/bingo) and built automatically on first use.

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
