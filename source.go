package rv1

import (
	"io/fs"

	"github.com/perdasilva/rv1/internal/bundle/source"
)

// BundleSource loads a registry+v1 bundle from some backing store.
type BundleSource = source.BundleSource

// FromFS creates a BundleSource that reads a registry+v1 bundle from
// the given filesystem. The filesystem should contain manifests/ and
// metadata/ directories in the standard OLM bundle layout.
func FromFS(fsys fs.FS) BundleSource {
	return source.FromFS(fsys)
}

// FromBundle creates a BundleSource from an already-parsed RegistryV1 bundle.
func FromBundle(rv1 RegistryV1) BundleSource {
	return source.FromBundle(rv1)
}
