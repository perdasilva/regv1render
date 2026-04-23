package render

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/regv1render/internal/bundle"
)

// DeploymentConfig is a type alias for v1alpha1.SubscriptionConfig
// to maintain clear naming in the OLMv1 context while reusing the v0 type.
type DeploymentConfig = v1alpha1.SubscriptionConfig

// BundleValidator validates a RegistryV1 bundle.
type BundleValidator interface {
	Validate(rv1 *bundle.RegistryV1) error
}

// ResourceGenerator generates resources given a registry+v1 bundle and options.
type ResourceGenerator func(rv1 *bundle.RegistryV1, opts Options) ([]client.Object, error)

func (g ResourceGenerator) GenerateResources(rv1 *bundle.RegistryV1, opts Options) ([]client.Object, error) {
	return g(rv1, opts)
}

// ResourceGenerators aggregates generators.
type ResourceGenerators []ResourceGenerator

func (r ResourceGenerators) GenerateResources(rv1 *bundle.RegistryV1, opts Options) ([]client.Object, error) {
	//nolint:prealloc
	var renderedObjects []client.Object
	for _, generator := range r {
		objs, err := generator.GenerateResources(rv1, opts)
		if err != nil {
			return nil, err
		}
		renderedObjects = append(renderedObjects, objs...)
	}
	return renderedObjects, nil
}

// UniqueNameGenerator produces deterministic unique names.
type UniqueNameGenerator func(string, interface{}) string

// Options holds the resolved configuration for a render call.
// This is an internal type passed to generators.
type Options struct {
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

func validateTargetNamespaces(rv1 *bundle.RegistryV1, installNamespace string, targetNamespaces []string) error {
	supportedInstallModes := supportedBundleInstallModes(rv1)

	set := sets.New[string](targetNamespaces...)
	switch {
	case set.Len() == 0:
		// Note: this function generally expects targetNamespace to contain at least one value set by default
		// in case the user does not specify the value. The option to set the targetNamespace is a no-op if it is empty.
		// The only case for which a default targetNamespace is undefined is in the case of a bundle that only
		// supports SingleNamespace install mode. The if statement here is added to provide a more friendly error
		// message than just the generic (at least one target namespace must be specified) which would occur
		// in case only the MultiNamespace install mode is supported by the bundle.
		// If AllNamespaces mode is supported, the default will be [""] -> watch all namespaces
		// If only OwnNamespace is supported, the default will be [install-namespace] -> only watch the install/own namespace
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
