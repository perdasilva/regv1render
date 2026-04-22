package generators

import (
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/regv1render/internal/bundle"
	"github.com/perdasilva/regv1render/internal/render"
)

var (
	adminVerbs = []string{"*"}
	editVerbs  = []string{"create", "update", "patch", "delete"}
	viewVerbs  = []string{"get", "list", "watch"}

	verbsForSuffix = map[string][]string{
		"admin": adminVerbs,
		"edit":  editVerbs,
		"view":  viewVerbs,
	}
)

func BundleProvidedAPIsClusterRolesGenerator(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
	objects := make([]client.Object, 0, len(rv1.CSV.Spec.CustomResourceDefinitions.Owned)*4)

	for _, owned := range rv1.CSV.Spec.CustomResourceDefinitions.Owned {
		nameGroupPair := strings.SplitN(owned.Name, ".", 2)
		if len(nameGroupPair) != 2 {
			return nil, fmt.Errorf("invalid CRD name %q: expected <plural>.<group>", owned.Name)
		}
		plural := nameGroupPair[0]
		group := nameGroupPair[1]
		namePrefix := fmt.Sprintf("%s-%s-", owned.Name, owned.Version)

		for suffix, verbs := range verbsForSuffix {
			objects = append(objects, &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: namePrefix + suffix,
					Labels: map[string]string{
						fmt.Sprintf("rbac.authorization.k8s.io/aggregate-to-%s", suffix): "true",
					},
				},
				Rules: []rbacv1.PolicyRule{{
					Verbs:     verbs,
					APIGroups: []string{group},
					Resources: []string{plural},
				}},
			})
		}

		objects = append(objects, &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix + "crd-view",
				Labels: map[string]string{
					"rbac.authorization.k8s.io/aggregate-to-view": "true",
				},
			},
			Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"get"},
				APIGroups:     []string{"apiextensions.k8s.io"},
				Resources:     []string{"customresourcedefinitions"},
				ResourceNames: []string{owned.Name},
			}},
		})
	}

	return objects, nil
}
