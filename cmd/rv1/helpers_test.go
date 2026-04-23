package main

import (
	"io/fs"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/rv1"
)

func fromFSHelper(t *testing.T, fsys fs.FS) rv1.BundleSource {
	t.Helper()
	return rv1.FromFS(fsys)
}

func renderBundle(source rv1.BundleSource, installNamespace string, watchNamespaces []string, cfg renderConfig) ([]client.Object, error) {
	rv1, err := source.GetBundle()
	if err != nil {
		return nil, err
	}
	renderer := buildRenderer(cfg)
	renderOpts := buildRenderOptions(cfg, watchNamespaces)
	return renderer.Render(rv1, installNamespace, renderOpts...)
}
