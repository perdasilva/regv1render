package regv1render_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/regv1render"
)

func TestWithProvidedAPIsClusterRoles(t *testing.T) {
	rv1 := regv1render.RegistryV1{
		PackageName: "test-operator",
		CSV: v1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{Name: "test-operator.v1.0.0"},
			Spec: v1alpha1.ClusterServiceVersionSpec{
				InstallModes: []v1alpha1.InstallMode{
					{Type: v1alpha1.InstallModeTypeAllNamespaces, Supported: true},
				},
				InstallStrategy: v1alpha1.NamedInstallStrategy{
					StrategyName: "deployment",
					StrategySpec: v1alpha1.StrategyDetailsDeployment{
						DeploymentSpecs: []v1alpha1.StrategyDeploymentSpec{{
							Name: "test-operator",
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

	t.Run("without option produces no provided API roles", func(t *testing.T) {
		objs, err := regv1render.Render(rv1, "test-ns")
		require.NoError(t, err)

		for _, obj := range objs {
			if cr, ok := obj.(*rbacv1.ClusterRole); ok {
				assert.NotContains(t, cr.Name, "widgets.example.com-v1-admin")
				assert.NotContains(t, cr.Name, "widgets.example.com-v1-edit")
				assert.NotContains(t, cr.Name, "widgets.example.com-v1-view")
			}
		}
	})

	t.Run("with option produces provided API roles", func(t *testing.T) {
		objs, err := regv1render.Render(rv1, "test-ns",
			regv1render.WithProvidedAPIsClusterRoles(),
		)
		require.NoError(t, err)

		roleNames := map[string]bool{}
		for _, obj := range objs {
			if cr, ok := obj.(*rbacv1.ClusterRole); ok {
				roleNames[cr.Name] = true
			}
		}

		assert.True(t, roleNames["widgets.example.com-v1-admin"])
		assert.True(t, roleNames["widgets.example.com-v1-edit"])
		assert.True(t, roleNames["widgets.example.com-v1-view"])
		assert.True(t, roleNames["widgets.example.com-v1-crd-view"])
	})
}
