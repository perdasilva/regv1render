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
// For more control, use the DefaultRenderer directly or construct a custom
// BundleRenderer with specific validators and generators.
//
// Bundles can be loaded from a filesystem using FromFS:
//
//	source := regv1render.FromFS(os.DirFS("path/to/bundle"))
//	rv1, err := source.GetBundle()
//
// OLMv0 compatibility features (such as provided API ClusterRoles) are
// available as opt-in rendering options:
//
//	objs, err := regv1render.Render(rv1, "ns", regv1render.WithProvidedAPIsClusterRoles())
package regv1render
