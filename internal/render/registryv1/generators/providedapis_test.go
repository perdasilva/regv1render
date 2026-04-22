package generators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/regv1render/internal/bundle"
	"github.com/perdasilva/regv1render/internal/render"
	"github.com/perdasilva/regv1render/internal/render/registryv1/generators"
)

func TestBundleProvidedAPIsClusterRolesGenerator(t *testing.T) {
	t.Run("single owned CRD generates 4 ClusterRoles", func(t *testing.T) {
		rv1 := &bundle.RegistryV1{
			CSV: v1alpha1.ClusterServiceVersion{
				Spec: v1alpha1.ClusterServiceVersionSpec{
					CustomResourceDefinitions: v1alpha1.CustomResourceDefinitions{
						Owned: []v1alpha1.CRDDescription{{
							Name:    "etcdclusters.etcd.database.coreos.com",
							Version: "v1beta2",
							Kind:    "EtcdCluster",
						}},
					},
				},
			},
		}

		objs, err := generators.BundleProvidedAPIsClusterRolesGenerator(rv1, render.Options{})
		require.NoError(t, err)
		require.Len(t, objs, 4)

		rolesByName := map[string]*rbacv1.ClusterRole{}
		for _, obj := range objs {
			cr := obj.(*rbacv1.ClusterRole)
			rolesByName[cr.Name] = cr
		}

		admin := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-admin"]
		require.NotNil(t, admin)
		assert.Equal(t, "true", admin.Labels["rbac.authorization.k8s.io/aggregate-to-admin"])
		assert.Equal(t, []rbacv1.PolicyRule{{
			Verbs:     []string{"*"},
			APIGroups: []string{"etcd.database.coreos.com"},
			Resources: []string{"etcdclusters"},
		}}, admin.Rules)

		edit := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-edit"]
		require.NotNil(t, edit)
		assert.Equal(t, "true", edit.Labels["rbac.authorization.k8s.io/aggregate-to-edit"])
		assert.Equal(t, []rbacv1.PolicyRule{{
			Verbs:     []string{"create", "update", "patch", "delete"},
			APIGroups: []string{"etcd.database.coreos.com"},
			Resources: []string{"etcdclusters"},
		}}, edit.Rules)

		view := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-view"]
		require.NotNil(t, view)
		assert.Equal(t, "true", view.Labels["rbac.authorization.k8s.io/aggregate-to-view"])
		assert.Equal(t, []rbacv1.PolicyRule{{
			Verbs:     []string{"get", "list", "watch"},
			APIGroups: []string{"etcd.database.coreos.com"},
			Resources: []string{"etcdclusters"},
		}}, view.Rules)

		crdView := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-crd-view"]
		require.NotNil(t, crdView)
		assert.Equal(t, "true", crdView.Labels["rbac.authorization.k8s.io/aggregate-to-view"])
		assert.Equal(t, []rbacv1.PolicyRule{{
			Verbs:         []string{"get"},
			APIGroups:     []string{"apiextensions.k8s.io"},
			Resources:     []string{"customresourcedefinitions"},
			ResourceNames: []string{"etcdclusters.etcd.database.coreos.com"},
		}}, crdView.Rules)
	})

	t.Run("multiple owned CRDs generate 4 ClusterRoles each", func(t *testing.T) {
		rv1 := &bundle.RegistryV1{
			CSV: v1alpha1.ClusterServiceVersion{
				Spec: v1alpha1.ClusterServiceVersionSpec{
					CustomResourceDefinitions: v1alpha1.CustomResourceDefinitions{
						Owned: []v1alpha1.CRDDescription{
							{Name: "foos.example.com", Version: "v1", Kind: "Foo"},
							{Name: "bars.example.com", Version: "v1", Kind: "Bar"},
							{Name: "bazzes.other.io", Version: "v2alpha1", Kind: "Baz"},
						},
					},
				},
			},
		}

		objs, err := generators.BundleProvidedAPIsClusterRolesGenerator(rv1, render.Options{})
		require.NoError(t, err)
		assert.Len(t, objs, 12)
	})

	t.Run("no owned CRDs generates no ClusterRoles", func(t *testing.T) {
		rv1 := &bundle.RegistryV1{
			CSV: v1alpha1.ClusterServiceVersion{
				Spec: v1alpha1.ClusterServiceVersionSpec{
					CustomResourceDefinitions: v1alpha1.CustomResourceDefinitions{
						Owned: []v1alpha1.CRDDescription{},
					},
				},
			},
		}

		objs, err := generators.BundleProvidedAPIsClusterRolesGenerator(rv1, render.Options{})
		require.NoError(t, err)
		assert.Empty(t, objs)
	})

	t.Run("invalid CRD name returns error", func(t *testing.T) {
		rv1 := &bundle.RegistryV1{
			CSV: v1alpha1.ClusterServiceVersion{
				Spec: v1alpha1.ClusterServiceVersionSpec{
					CustomResourceDefinitions: v1alpha1.CustomResourceDefinitions{
						Owned: []v1alpha1.CRDDescription{{
							Name:    "invalidname",
							Version: "v1",
							Kind:    "Invalid",
						}},
					},
				},
			},
		}

		_, err := generators.BundleProvidedAPIsClusterRolesGenerator(rv1, render.Options{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid CRD name")
	})
}
