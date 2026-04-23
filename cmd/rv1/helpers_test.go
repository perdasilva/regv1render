package main

import (
	"io/fs"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/regv1render"
)

func fromFSHelper(t *testing.T, fsys fs.FS) regv1render.BundleSource {
	t.Helper()
	return regv1render.FromFS(fsys)
}

func renderBundle(source regv1render.BundleSource, installNamespace string, watchNamespaces []string, cfg renderConfig) ([]client.Object, error) {
	rv1, err := source.GetBundle()
	if err != nil {
		return nil, err
	}
	renderer := buildRenderer(cfg)
	renderOpts := buildRenderOptions(cfg, watchNamespaces)
	return renderer.Render(rv1, installNamespace, renderOpts...)
}
