package regv1render

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/regv1render/internal/bundle"
	"github.com/perdasilva/regv1render/internal/render"
	"github.com/perdasilva/regv1render/internal/render/registryv1"
	"github.com/perdasilva/regv1render/internal/render/registryv1/generators"
)

// BundleRenderer validates and renders a registry+v1 bundle into plain
// Kubernetes manifests. Use DefaultRenderer for a pre-configured instance
// with all standard validators and generators.
type BundleRenderer = render.BundleRenderer

// BundleValidator validates a registry+v1 bundle for correctness
// before rendering.
type BundleValidator = render.BundleValidator

// ValidationError represents a validation failure from a specific check.
// Use errors.As to extract it from validation errors and inspect the
// Check field to identify which validation rule failed.
type ValidationError = render.ValidationError

// ResourceGenerator is a function that generates Kubernetes resources
// from a registry+v1 bundle and rendering options.
type ResourceGenerator = render.ResourceGenerator

// ResourceGenerators is a list of ResourceGenerator functions that can
// be composed into a single generator.
type ResourceGenerators = render.ResourceGenerators

// Options configures how a bundle is rendered, including target
// namespaces, certificate providers, and deployment configuration.
type Options = render.Options

// Option is a functional option for configuring rendering.
type Option = render.Option

// UniqueNameGenerator produces deterministic unique names for generated
// resources based on a base name and an input object.
type UniqueNameGenerator = render.UniqueNameGenerator

// CertificateProvider is an interface for injecting TLS certificate
// management into rendered webhook and API service resources.
type CertificateProvider = render.CertificateProvider

// CertSecretInfo contains the name and namespace of a TLS secret
// used by a certificate provider.
type CertSecretInfo = render.CertSecretInfo

// CertificateProvisioner is a concrete CertificateProvider built from
// a CertificateProvisionerConfig.
type CertificateProvisioner = render.CertificateProvisioner

// CertificateProvisionerConfig holds the configuration for building
// a CertificateProvisioner.
type CertificateProvisionerConfig = render.CertificateProvisionerConfig

// RegistryV1 holds the parsed contents of a registry+v1 bundle,
// including its ClusterServiceVersion, CRDs, and other objects.
type RegistryV1 = bundle.RegistryV1

// DeploymentConfig allows customizing the deployment resources
// generated during rendering (node selectors, tolerations, resources,
// env vars, volumes, affinity, annotations).
type DeploymentConfig = render.DeploymentConfig

// DefaultRenderer is a pre-configured BundleRenderer with all standard
// validators and resource generators for registry+v1 bundles.
var DefaultRenderer = registryv1.Renderer

// WithTargetNamespaces sets the namespaces the operator should watch.
func WithTargetNamespaces(namespaces ...string) Option {
	return render.WithTargetNamespaces(namespaces...)
}

// WithUniqueNameGenerator sets a custom name generator for rendered resources.
func WithUniqueNameGenerator(generator UniqueNameGenerator) Option {
	return render.WithUniqueNameGenerator(generator)
}

// WithCertificateProvider sets the certificate provider for webhook TLS.
func WithCertificateProvider(provider CertificateProvider) Option {
	return render.WithCertificateProvider(provider)
}

// WithDeploymentConfig sets deployment customization options.
func WithDeploymentConfig(deploymentConfig *DeploymentConfig) Option {
	return render.WithDeploymentConfig(deploymentConfig)
}

// WithProvidedAPIsClusterRoles enables generation of aggregated
// admin/edit/view ClusterRoles for each owned CRD, matching the
// OLMv0 (operator-lifecycle-manager) behavior. This is opt-in and
// does not affect default rendering.
func WithProvidedAPIsClusterRoles() Option {
	return func(o *render.Options) {
		o.AdditionalGenerators = append(o.AdditionalGenerators, generators.BundleProvidedAPIsClusterRolesGenerator)
	}
}

// DefaultUniqueNameGenerator produces deterministic unique names by
// hashing the input object and appending it to the base name.
func DefaultUniqueNameGenerator(base string, o interface{}) string {
	return render.DefaultUniqueNameGenerator(base, o)
}

// CertProvisionerFor creates a CertificateProvisioner configured for
// the given deployment name and rendering options.
func CertProvisionerFor(deploymentName string, opts Options) CertificateProvisioner {
	return render.CertProvisionerFor(deploymentName, opts)
}

// Render is a convenience function that renders a registry+v1 bundle
// using the DefaultRenderer.
func Render(rv1 RegistryV1, installNamespace string, opts ...Option) ([]client.Object, error) {
	return DefaultRenderer.Render(rv1, installNamespace, opts...)
}
