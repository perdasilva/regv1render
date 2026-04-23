package registryv1_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/regv1render/internal/bundle"
	"github.com/perdasilva/regv1render/internal/render/registryv1"
	. "github.com/perdasilva/regv1render/internal/util/testutil"
	"github.com/perdasilva/regv1render/internal/util/testutil/clusterserviceversion"
)

func Test_Renderer_Success(t *testing.T) {
	someBundle := bundle.RegistryV1{
		PackageName: "my-package",
		CSV: clusterserviceversion.Builder().
			WithName("test-bundle").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		Others: []unstructured.Unstructured{
			*ToUnstructuredT(t, &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-service",
				},
			}),
		},
	}

	renderer := registryv1.NewRendererBuilder().Build()
	objs, err := renderer.Render(someBundle, "install-namespace")
	require.NoError(t, err)
	require.NotEmpty(t, objs)
	require.Len(t, objs, 1)
	require.Equal(t, "my-service", objs[0].GetName())
	require.Equal(t, "install-namespace", objs[0].GetNamespace())
}

func Test_Renderer_Failure_UnsupportedKind(t *testing.T) {
	someBundle := bundle.RegistryV1{
		PackageName: "my-package",
		CSV: clusterserviceversion.Builder().
			WithName("test-bundle").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		Others: []unstructured.Unstructured{
			*ToUnstructuredT(t, &corev1.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testEvent",
				},
			}),
		},
	}

	renderer := registryv1.NewRendererBuilder().Build()
	objs, err := renderer.Render(someBundle, "install-namespace")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported resource")
	require.Empty(t, objs)
}
