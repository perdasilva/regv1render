package render

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/rv1/internal/bundle"
	"github.com/perdasilva/rv1/internal/render/validator"
)

// DeploymentConfig is a type alias for v1alpha1.SubscriptionConfig
// to maintain clear naming in the OLMv1 context while reusing the v0 type.
type DeploymentConfig = v1alpha1.SubscriptionConfig

// bundleValidator validates a RegistryV1 bundle.
type bundleValidator interface {
	Validate(rv1 *bundle.RegistryV1) error
}

// resourceGenerator generates resources given a registry+v1 bundle and options.
type resourceGenerator func(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error)

// UniqueNameGenerator produces deterministic unique names.
type UniqueNameGenerator func(string, interface{}) string

// options holds the resolved configuration for a render call.
type options struct {
	InstallNamespace         string
	TargetNamespaces         []string
	UniqueNameGenerator      UniqueNameGenerator
	CertificateProvider      CertificateProvider
	DeploymentConfig         *DeploymentConfig
	ProvidedAPIsClusterRoles bool
}

func DefaultUniqueNameGenerator(base string, o interface{}) string {
	hashStr := DeepHashObject(o)
	return ObjectNameForBaseAndSuffix(base, hashStr)
}

// Renderer validates and renders registry+v1 bundles to plain Kubernetes manifests.
type Renderer struct {
	validator           bundleValidator
	generators          []resourceGenerator
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

// NewRendererBuilder creates a RendererBuilder with the standard registry+v1
// validator and generators.
func NewRendererBuilder() *RendererBuilder {
	return &RendererBuilder{
		renderer: Renderer{
			validator:           validator.BundleValidator{},
			generators:          defaultGenerators(),
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

	genOpts := options{
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

	//nolint:prealloc
	var renderedObjects []client.Object
	for _, generator := range r.generators {
		objs, err := generator(&rv1, genOpts)
		if err != nil {
			return nil, err
		}
		renderedObjects = append(renderedObjects, objs...)
	}

	return renderedObjects, nil
}

func validateTargetNamespaces(rv1 *bundle.RegistryV1, installNamespace string, targetNamespaces []string) error {
	supportedInstallModes := supportedBundleInstallModes(rv1)

	set := sets.New[string](targetNamespaces...)
	switch {
	case set.Len() == 0:
		if supportedInstallModes.Has(v1alpha1.InstallModeTypeMultiNamespace) {
			return errors.New("at least one target namespace must be specified")
		}
		return errors.New("exactly one target namespace must be specified")
	case set.Len() == 1 && set.Has(""):
		if supportedInstallModes.Has(v1alpha1.InstallModeTypeAllNamespaces) {
			return nil
		}
		return fmt.Errorf("supported install modes %v do not support targeting all namespaces", sets.List(supportedInstallModes))
	case set.Len() == 1 && !set.Has(""):
		if targetNamespaces[0] == installNamespace {
			if !supportedInstallModes.Has(v1alpha1.InstallModeTypeOwnNamespace) {
				return fmt.Errorf("supported install modes %v do not support targeting own namespace", sets.List(supportedInstallModes))
			}
			return nil
		}
		if supportedInstallModes.Has(v1alpha1.InstallModeTypeSingleNamespace) {
			return nil
		}
	default:
		if !supportedInstallModes.Has(v1alpha1.InstallModeTypeOwnNamespace) && set.Has(installNamespace) {
			return fmt.Errorf("supported install modes %v do not support targeting own namespace", sets.List(supportedInstallModes))
		}
		if supportedInstallModes.Has(v1alpha1.InstallModeTypeMultiNamespace) && !set.Has("") {
			return nil
		}
	}
	return fmt.Errorf("supported install modes %v do not support target namespaces %v", sets.List[v1alpha1.InstallModeType](supportedInstallModes), targetNamespaces)
}

func defaultTargetNamespacesForBundle(rv1 *bundle.RegistryV1) []string {
	supportedInstallModes := supportedBundleInstallModes(rv1)

	if supportedInstallModes.Has(v1alpha1.InstallModeTypeAllNamespaces) {
		return []string{corev1.NamespaceAll}
	}

	return nil
}

func supportedBundleInstallModes(rv1 *bundle.RegistryV1) sets.Set[v1alpha1.InstallModeType] {
	supportedInstallModes := sets.New[v1alpha1.InstallModeType]()
	for _, im := range rv1.CSV.Spec.InstallModes {
		if im.Supported {
			supportedInstallModes.Insert(im.Type)
		}
	}
	return supportedInstallModes
}
