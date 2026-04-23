// Package rv1 renders OLM registry+v1 bundles to plain Kubernetes manifests.
//
// This library is extracted from the operator-framework/operator-controller rendering
// pipeline and is compatible with the operator-framework/operator-lifecycle-manager
// rendering behavior.
//
// The simplest way to render a bundle is with the top-level Render function:
//
//	objs, err := rv1.Render(bundle, "install-namespace")
//
// For more control, use the builder to create a configured Renderer:
//
//	r := rv1.NewRendererBuilder().
//	    WithCertificateProvider(rv1.CertManagerProvider{}).
//	    Build()
//	objs, err := r.Render(bundle, "install-namespace",
//	    rv1.WithTargetNamespaces("watch-ns"),
//	    rv1.WithProvidedAPIsClusterRoles(),
//	)
//
// Bundles can be loaded from a filesystem using FromFS:
//
//	source := rv1.FromFS(os.DirFS("path/to/bundle"))
//	bundle, err := source.GetBundle()
package rv1
