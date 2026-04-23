package registryv1

import (
	"github.com/perdasilva/regv1render/internal/render"
	"github.com/perdasilva/regv1render/internal/render/registryv1/generators"
	"github.com/perdasilva/regv1render/internal/render/registryv1/validator"
)

// Generators is the standard set of resource generators for registry+v1 bundles.
var Generators = []render.ResourceGenerator{
	generators.BundleCSVServiceAccountGenerator,
	generators.BundleCSVPermissionsGenerator,
	generators.BundleCSVClusterPermissionsGenerator,
	generators.BundleCRDGenerator,
	generators.BundleAdditionalResourcesGenerator,
	generators.BundleCSVDeploymentGenerator,
	generators.BundleValidatingWebhookResourceGenerator,
	generators.BundleMutatingWebhookResourceGenerator,
	generators.BundleDeploymentServiceResourceGenerator,
	generators.CertProviderResourceGenerator,
	generators.BundleProvidedAPIsClusterRolesGenerator,
}

// NewRendererBuilder creates a RendererBuilder with the standard registry+v1
// validator and generators.
func NewRendererBuilder() *render.RendererBuilder {
	return render.NewRendererBuilder(validator.BundleValidator{}, Generators)
}
