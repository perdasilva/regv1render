package certproviders_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/rv1/internal/render/certproviders"
	"github.com/perdasilva/rv1/internal/renderutil"
)

func toUnstructuredHelper(t *testing.T, obj client.Object) *unstructured.Unstructured {
	t.Helper()
	u, err := renderutil.ToUnstructured(obj)
	require.NoError(t, err)
	return u
}

func Test_CertificateProvisioner_WithoutCertProvider(t *testing.T) {
	provisioner := &certproviders.CertificateProvisioner{
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
	mockProvider := certproviders.NewMockCertificateProvider(t)
	mockProvider.EXPECT().InjectCABundle(mock.Anything, mock.Anything).
		RunAndReturn(func(obj client.Object, cfg certproviders.CertificateProvisionerConfig) error {
			obj.SetName("some-name")
			return nil
		})
	mockProvider.EXPECT().AdditionalObjects(mock.Anything).
		Return([]unstructured.Unstructured{*toUnstructuredHelper(t, &corev1.Secret{
			TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()},
		})}, nil)
	mockProvider.EXPECT().GetCertSecretInfo(mock.Anything).
		Return(certproviders.CertSecretInfo{
			SecretName:     "some-secret",
			PrivateKeyKey:  "some-key",
			CertificateKey: "another-key",
		})
	provisioner := &certproviders.CertificateProvisioner{
		ServiceName:  "webhook",
		CertName:     "cert",
		Namespace:    "namespace",
		CertProvider: mockProvider,
	}

	svc := &corev1.Service{}
	require.NoError(t, provisioner.InjectCABundle(svc))
	require.Equal(t, "some-name", svc.GetName())

	objs, err := provisioner.AdditionalObjects()
	require.NoError(t, err)
	require.Equal(t, []unstructured.Unstructured{*toUnstructuredHelper(t, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()},
	})}, objs)

	require.Equal(t, &certproviders.CertSecretInfo{
		SecretName:     "some-secret",
		PrivateKeyKey:  "some-key",
		CertificateKey: "another-key",
	}, provisioner.GetCertSecretInfo())
}

func Test_CertificateProvisioner_Errors(t *testing.T) {
	mockProvider := certproviders.NewMockCertificateProvider(t)
	mockProvider.EXPECT().InjectCABundle(mock.Anything, mock.Anything).
		Return(fmt.Errorf("some error"))
	mockProvider.EXPECT().AdditionalObjects(mock.Anything).
		Return(nil, fmt.Errorf("some other error"))
	provisioner := &certproviders.CertificateProvisioner{
		ServiceName:  "webhook",
		CertName:     "cert",
		Namespace:    "namespace",
		CertProvider: mockProvider,
	}

	err := provisioner.InjectCABundle(&corev1.Service{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "some error")

	objs, err := provisioner.AdditionalObjects()
	require.Error(t, err)
	require.Contains(t, err.Error(), "some other error")
	require.Nil(t, objs)
}
