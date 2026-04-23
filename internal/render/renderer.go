package render

import (
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/regv1render/internal/bundle"
)

// Renderer validates and renders registry+v1 bundles to plain Kubernetes manifests.
type Renderer struct {
	validator           BundleValidator
	generators          []ResourceGenerator
	certProvider        CertificateProvider
	uniqueNameGenerator UniqueNameGenerator
	deploymentConfig    *DeploymentConfig
}

// RendererBuilder constructs a Renderer using fluent method chaining.
type RendererBuilder struct {
	renderer Renderer
}

// RenderOption configures a single Render call.
type RenderOption func(*renderCallOptions)

type renderCallOptions struct {
	targetNamespaces         []string
	providedAPIsClusterRoles bool
}

// NewRendererBuilder creates a RendererBuilder with the given validator and generators.
func NewRendererBuilder(validator BundleValidator, generators []ResourceGenerator) *RendererBuilder {
	return &RendererBuilder{
		renderer: Renderer{
			validator:           validator,
			generators:          generators,
			uniqueNameGenerator: DefaultUniqueNameGenerator,
		},
	}
}

// WithCertificateProvider sets the certificate provider for webhook TLS.
func (b *RendererBuilder) WithCertificateProvider(provider CertificateProvider) *RendererBuilder {
	b.renderer.certProvider = provider
	return b
}

// WithDeploymentConfig sets deployment customization options.
func (b *RendererBuilder) WithDeploymentConfig(deploymentConfig *DeploymentConfig) *RendererBuilder {
	b.renderer.deploymentConfig = deploymentConfig
	return b
}

// WithUniqueNameGenerator sets a custom name generator for rendered resources.
func (b *RendererBuilder) WithUniqueNameGenerator(generator UniqueNameGenerator) *RendererBuilder {
	b.renderer.uniqueNameGenerator = generator
	return b
}

// Build creates the Renderer from the builder configuration.
func (b *RendererBuilder) Build() *Renderer {
	r := b.renderer
	return &r
}

// WithProvidedAPIsClusterRoles enables generation of aggregated
// admin/edit/view ClusterRoles for each owned CRD.
func WithProvidedAPIsClusterRoles() RenderOption {
	return func(o *renderCallOptions) {
		o.providedAPIsClusterRoles = true
	}
}

// WithTargetNamespaces sets the namespaces the operator should watch.
func WithTargetNamespaces(namespaces ...string) RenderOption {
	return func(o *renderCallOptions) {
		if len(namespaces) > 0 {
			o.targetNamespaces = namespaces
		}
	}
}

// Render validates and renders a registry+v1 bundle.
func (r *Renderer) Render(rv1 bundle.RegistryV1, installNamespace string, opts ...RenderOption) ([]client.Object, error) {
	if r.validator != nil {
		if err := r.validator.Validate(&rv1); err != nil {
			return nil, err
		}
	}

	callOpts := &renderCallOptions{}
	for _, opt := range opts {
		opt(callOpts)
	}

	targetNamespaces := callOpts.targetNamespaces
	if len(targetNamespaces) == 0 {
		targetNamespaces = defaultTargetNamespacesForBundle(&rv1)
	}

	genOpts := Options{
		InstallNamespace:         installNamespace,
		TargetNamespaces:         targetNamespaces,
		UniqueNameGenerator:      r.uniqueNameGenerator,
		CertificateProvider:      r.certProvider,
		DeploymentConfig:         r.deploymentConfig,
		ProvidedAPIsClusterRoles: callOpts.providedAPIsClusterRoles,
	}

	if genOpts.UniqueNameGenerator == nil {
		return nil, errors.New("unique name generator must be specified")
	}
	if err := validateTargetNamespaces(&rv1, installNamespace, genOpts.TargetNamespaces); err != nil {
		return nil, fmt.Errorf("invalid target namespaces %v: %w", genOpts.TargetNamespaces, err)
	}

	objs, err := ResourceGenerators(r.generators).GenerateResources(&rv1, genOpts)
	if err != nil {
		return nil, err
	}

	return objs, nil
}
