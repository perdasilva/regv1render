// Package rv1 renders OLM registry+v1 bundles to plain Kubernetes manifests.
//
// This library is extracted from the operator-framework/operator-controller rendering
// pipeline and is compatible with the operator-framework/operator-lifecycle-manager
// rendering behavior.
//
// Create a Renderer using the builder:
//
//	renderer := rv1.NewRendererBuilder().Build()
//	objs, err := renderer.Render(bundle, "install-namespace")
//
// Configure certificate providers and render options:
//
//	renderer := rv1.NewRendererBuilder().
//	    WithCertificateProvider(rv1.CertManagerProvider{}).
//	    Build()
//	objs, err := renderer.Render(bundle, "install-namespace",
//	    rv1.WithTargetNamespaces("watch-ns"),
//	    rv1.WithProvidedAPIsClusterRoles(),
//	)
//
// Bundles can be loaded from a filesystem using FromFS:
//
//	source := rv1.FromFS(os.DirFS("path/to/bundle"))
//	bundle, err := source.GetBundle()
package rv1
