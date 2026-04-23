// Package regv1render renders OLM registry+v1 bundles to plain Kubernetes manifests.
//
// This library is extracted from the operator-framework/operator-controller rendering
// pipeline and is compatible with the operator-framework/operator-lifecycle-manager
// rendering behavior.
//
// The simplest way to render a bundle is with the top-level Render function:
//
//	objs, err := regv1render.Render(rv1, "install-namespace")
//
// For more control, use the builder to create a configured Renderer:
//
//	r := regv1render.NewRendererBuilder().
//	    WithCertificateProvider(regv1render.CertManagerProvider{}).
//	    Build()
//	objs, err := r.Render(rv1, "install-namespace",
//	    regv1render.WithTargetNamespaces("watch-ns"),
//	    regv1render.WithProvidedAPIsClusterRoles(),
//	)
//
// Bundles can be loaded from a filesystem using FromFS:
//
//	source := regv1render.FromFS(os.DirFS("path/to/bundle"))
//	rv1, err := source.GetBundle()
package regv1render
