package registryv1

import (
	"github.com/perdasilva/regv1render/internal/render"
	"github.com/perdasilva/regv1render/internal/render/registryv1/generators"
	"github.com/perdasilva/regv1render/internal/render/registryv1/validator"
)

// Renderer renders registry+v1 bundles into plain kubernetes manifests
var Renderer = render.BundleRenderer{
	BundleValidator:    validator.BundleValidator{},
	ResourceGenerators: ResourceGenerators,
}

// ResourceGenerators a slice of ResourceGenerators required to generate plain resource manifests for
// registry+v1 bundles
var ResourceGenerators = []render.ResourceGenerator{
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
}
