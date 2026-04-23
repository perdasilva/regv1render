package rv1

import (
	"github.com/perdasilva/rv1/internal/bundle"
	"github.com/perdasilva/rv1/internal/render"
	"github.com/perdasilva/rv1/internal/render/certproviders"
	"github.com/perdasilva/rv1/internal/render/validator"
)

// Renderer validates and renders registry+v1 bundles to plain
// Kubernetes manifests. Create one with NewRendererBuilder().Build().
type Renderer = render.Renderer

// RenderOption configures a single Render call.
type RenderOption = render.RenderOption

// CertificateProvider is an interface for injecting TLS certificate
// management into rendered webhook and API service resources.
type CertificateProvider = render.CertificateProvider

// DeploymentConfig allows customizing the deployment resources
// generated during rendering.
type DeploymentConfig = render.DeploymentConfig

// ValidationError represents a validation failure from a specific check.
type ValidationError = validator.ValidationError

// RegistryV1 holds the parsed contents of a registry+v1 bundle.
type RegistryV1 = bundle.RegistryV1

// CertManagerProvider is a CertificateProvider that uses cert-manager.
type CertManagerProvider = certproviders.CertManagerCertificateProvider

// OpenShiftServiceCAProvider is a CertificateProvider that uses OpenShift service-ca.
type OpenShiftServiceCAProvider = certproviders.OpenshiftServiceCaCertificateProvider

// SecretCertProvider is a CertificateProvider that generates a
// kubernetes.io/tls Secret for webhook TLS. If Cert and Key are
// empty, the Secret is created with empty data so users can
// populate it externally (Vault, manual, etc.).
type SecretCertProvider = certproviders.SecretCertProvider

// RendererBuilder constructs a Renderer using fluent method chaining.
type RendererBuilder struct {
	inner *render.RendererBuilder
}

// NewRendererBuilder creates a RendererBuilder with the standard
// registry+v1 validator and generators.
//
//	r := rv1.NewRendererBuilder().
//	    WithCertificateProvider(rv1.CertManagerProvider{}).
//	    WithProvidedAPIsClusterRoles().
//	    Build()
func NewRendererBuilder() *RendererBuilder {
	return &RendererBuilder{inner: render.NewRendererBuilder()}
}

// WithCertificateProvider sets the certificate provider for webhook TLS.
func (b *RendererBuilder) WithCertificateProvider(provider CertificateProvider) *RendererBuilder {
	b.inner.WithCertificateProvider(provider)
	return b
}

// WithDeploymentConfig sets deployment customization options.
func (b *RendererBuilder) WithDeploymentConfig(deploymentConfig *DeploymentConfig) *RendererBuilder {
	b.inner.WithDeploymentConfig(deploymentConfig)
	return b
}

// WithUniqueNameGenerator sets a custom name generator for rendered resources.
func (b *RendererBuilder) WithUniqueNameGenerator(generator func(string, interface{}) string) *RendererBuilder {
	b.inner.WithUniqueNameGenerator(generator)
	return b
}

// Build creates the Renderer from the builder configuration.
func (b *RendererBuilder) Build() *Renderer {
	return b.inner.Build()
}

// WithProvidedAPIsClusterRoles enables generation of aggregated
// admin/edit/view ClusterRoles for each owned CRD, matching
// OLMv0 behavior.
func WithProvidedAPIsClusterRoles() RenderOption {
	return render.WithProvidedAPIsClusterRoles()
}

// WithTargetNamespaces sets the namespaces the operator should watch.
func WithTargetNamespaces(namespaces ...string) RenderOption {
	return render.WithTargetNamespaces(namespaces...)
}
