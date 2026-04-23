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
certificateProvider:
  type: cert-manager  # or: openshift-service-ca, none
deploymentConfig:
  nodeSelector:
    kubernetes.io/os: linux
```

## OLMv0 Compatibility

The library includes opt-in support for rendering behaviors from the original [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager) (OLMv0).

### Provided API ClusterRoles

Use `WithProvidedAPIsClusterRoles()` to generate aggregated admin/edit/view ClusterRoles for each owned CRD, matching OLMv0 behavior:

```go
objs, err := regv1render.Render(rv1, "my-namespace",
    regv1render.WithProvidedAPIsClusterRoles(),
)
```

This generates 4 ClusterRoles per owned CRD:
- `<name>-<version>-admin` — full access (`*`)
- `<name>-<version>-edit` — write access (`create`, `update`, `patch`, `delete`)
- `<name>-<version>-view` — read access (`get`, `list`, `watch`)
- `<name>-<version>-crd-view` — read access to the CRD definition itself

These roles use Kubernetes [aggregated ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) so they're automatically included in the built-in admin/edit/view roles.

## Upstream Relationship

This library extracts and consolidates the registry+v1 bundle rendering logic from two operator-framework projects:

- **[operator-controller](https://github.com/operator-framework/operator-controller)** — the OLMv1 rendering pipeline (`internal/rukpak/render`), which provides the core `BundleRenderer`, validators, and resource generators
- **[operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)** — the OLMv0 rendering behaviors, available as opt-in options (e.g., provided API ClusterRoles)

The goal is upstream fidelity — rendering output matches the upstream implementations for the same inputs. The library includes regression tests with golden-file fixtures to verify this.

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

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full contributor guide.

## License

Apache-2.0
