package render

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func toUnstructuredHelper(t *testing.T, obj client.Object) *unstructured.Unstructured {
	t.Helper()
	u, err := ToUnstructured(obj)
	require.NoError(t, err)
	return u
}

func Test_CertificateProvisioner_WithoutCertProvider(t *testing.T) {
	provisioner := &CertificateProvisioner{
		ServiceName:  "webhook",
		CertName:     "cert",
		Namespace:    "namespace",
		CertProvider: nil,
	}

	require.NoError(t, provisioner.InjectCABundle(&corev1.Secret{}))
	require.Nil(t, provisioner.GetCertSecretInfo())

	objs, err := provisioner.AdditionalObjects()
	require.Nil(t, objs)
	require.NoError(t, err)
}

func Test_CertificateProvisioner_WithCertProvider(t *testing.T) {
	fakeProvider := &FakeCertProvider{
		InjectCABundleFn: func(obj client.Object, cfg CertificateProvisionerConfig) error {
			obj.SetName("some-name")
			return nil
		},
		AdditionalObjectsFn: func(cfg CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return []unstructured.Unstructured{*toUnstructuredHelper(t, &corev1.Secret{
				TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()},
			})}, nil
		},
		GetCertSecretInfoFn: func(cfg CertificateProvisionerConfig) CertSecretInfo {
			return CertSecretInfo{
				SecretName:     "some-secret",
				PrivateKeyKey:  "some-key",
				CertificateKey: "another-key",
			}
		},
	}
	provisioner := &CertificateProvisioner{
		ServiceName:  "webhook",
		CertName:     "cert",
		Namespace:    "namespace",
		CertProvider: fakeProvider,
	}

	svc := &corev1.Service{}
	require.NoError(t, provisioner.InjectCABundle(svc))
	require.Equal(t, "some-name", svc.GetName())

	objs, err := provisioner.AdditionalObjects()
	require.NoError(t, err)
	require.Equal(t, []unstructured.Unstructured{*toUnstructuredHelper(t, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()},
	})}, objs)

	require.Equal(t, &CertSecretInfo{
		SecretName:     "some-secret",
		PrivateKeyKey:  "some-key",
		CertificateKey: "another-key",
	}, provisioner.GetCertSecretInfo())
}

func Test_CertificateProvisioner_Errors(t *testing.T) {
	fakeProvider := &FakeCertProvider{
		InjectCABundleFn: func(obj client.Object, cfg CertificateProvisionerConfig) error {
			return fmt.Errorf("some error")
		},
		AdditionalObjectsFn: func(cfg CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return nil, fmt.Errorf("some other error")
		},
	}
	provisioner := &CertificateProvisioner{
		ServiceName:  "webhook",
		CertName:     "cert",
		Namespace:    "namespace",
		CertProvider: fakeProvider,
	}

	err := provisioner.InjectCABundle(&corev1.Service{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "some error")

	objs, err := provisioner.AdditionalObjects()
	require.Error(t, err)
	require.Contains(t, err.Error(), "some other error")
	require.Nil(t, objs)
}

func Test_CertProvisionerFor(t *testing.T) {
	fakeProvider := &FakeCertProvider{}
	prov := CertProvisionerFor("my.deployment.thing", options{
		InstallNamespace:    "my-namespace",
		CertificateProvider: fakeProvider,
	})

	require.Equal(t, prov.CertProvider, fakeProvider)
	require.Equal(t, "my-deployment-thing-service", prov.ServiceName)
	require.Equal(t, "my-deployment-thing-service-cert", prov.CertName)
	require.Equal(t, "my-namespace", prov.Namespace)
}

func Test_CertProvisionerFor_ExtraLargeName_MoreThan63Chars(t *testing.T) {
	prov := CertProvisionerFor("my.object.thing.has.a.really.really.really.really.really.long.name", options{})

	require.Len(t, prov.ServiceName, 63)
	require.Len(t, prov.CertName, 63)
	require.Equal(t, "my-object-thing-has-a-really-really-really-really-reall-service", prov.ServiceName)
	require.Equal(t, "my-object-thing-has-a-really-really-really-really-reall-se-cert", prov.CertName)
}
