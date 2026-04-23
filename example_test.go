package rv1_test

import (
	"fmt"
	"testing/fstest"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	regv1 "github.com/perdasilva/rv1"
)

func exampleBundle() regv1.RegistryV1 {
	return regv1.RegistryV1{
		PackageName: "my-operator",
		CSV: v1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{Name: "my-operator.v1.0.0"},
			Spec: v1alpha1.ClusterServiceVersionSpec{
				InstallModes: []v1alpha1.InstallMode{
					{Type: v1alpha1.InstallModeTypeAllNamespaces, Supported: true},
					{Type: v1alpha1.InstallModeTypeOwnNamespace, Supported: true},
					{Type: v1alpha1.InstallModeTypeSingleNamespace, Supported: true},
				},
				InstallStrategy: v1alpha1.NamedInstallStrategy{
					StrategyName: "deployment",
					StrategySpec: v1alpha1.StrategyDetailsDeployment{
						DeploymentSpecs: []v1alpha1.StrategyDeploymentSpec{{
							Name: "my-operator-controller",
						}},
					},
				},
				CustomResourceDefinitions: v1alpha1.CustomResourceDefinitions{
					Owned: []v1alpha1.CRDDescription{{
						Name:    "widgets.example.com",
						Version: "v1",
						Kind:    "Widget",
					}},
				},
			},
		},
		CRDs: []apiextensionsv1.CustomResourceDefinition{{
			ObjectMeta: metav1.ObjectMeta{Name: "widgets.example.com"},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "example.com",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "widgets", Singular: "widget", Kind: "Widget",
				},
				Scope:    apiextensionsv1.NamespaceScoped,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{Name: "v1", Served: true, Storage: true}},
			},
		}},
	}
}

func ExampleRender() {
	rv1 := exampleBundle()

	objs, err := regv1.Render(rv1, "operators")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Rendered %d objects\n", len(objs))
	// Output:
	// Rendered 2 objects
}

func ExampleRender_withTargetNamespaces() {
	rv1 := exampleBundle()

	objs, err := regv1.Render(rv1, "operators",
		regv1.WithTargetNamespaces("watch-ns"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Rendered %d objects\n", len(objs))
	// Output:
	// Rendered 2 objects
}

func ExampleRender_withProvidedAPIsClusterRoles() {
	rv1 := exampleBundle()

	objs, err := regv1.Render(rv1, "operators",
		regv1.WithProvidedAPIsClusterRoles(),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Rendered %d objects (includes provided API ClusterRoles)\n", len(objs))
	// Output:
	// Rendered 6 objects (includes provided API ClusterRoles)
}

func ExampleFromFS() {
	// Create an in-memory filesystem representing a bundle
	// In practice, this would be os.DirFS("path/to/bundle")
	bundleFS := fstest.MapFS{
		"metadata/annotations.yaml": &fstest.MapFile{
			Data: []byte(`annotations:
  operators.operatorframework.io.bundle.package.v1: my-operator
`),
		},
	}

	source := regv1.FromFS(bundleFS)
	_, err := source.GetBundle()

	// This will fail because the bundle is incomplete (no CSV),
	// but it demonstrates the FromFS API
	if err != nil {
		fmt.Println("FromFS loaded bundle source (bundle parsing requires a valid CSV)")
	}
	// Output:
	// FromFS loaded bundle source (bundle parsing requires a valid CSV)
}

func ExampleFromBundle() {
	rv1 := exampleBundle()

	// Wrap an already-parsed bundle as a BundleSource
	source := regv1.FromBundle(rv1)
	bundle, err := source.GetBundle()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bundle: %s\n", bundle.PackageName)
	fmt.Printf("CSV: %s\n", bundle.CSV.Name)
	// Output:
	// Bundle: my-operator
	// CSV: my-operator.v1.0.0
}

func ExampleRendererBuilder() {
	rv1 := exampleBundle()

	// Use NewRendererBuilder for full control
	objs, err := regv1.NewRendererBuilder().Build().Render(rv1, "operators")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Rendered %d objects using DefaultRenderer\n", len(objs))
	// Output:
	// Rendered 2 objects using DefaultRenderer
}
