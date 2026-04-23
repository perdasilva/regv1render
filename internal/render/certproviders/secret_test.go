package certproviders_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/perdasilva/regv1render/internal/render"
	"github.com/perdasilva/regv1render/internal/render/certproviders"
)

func TestSecretCertProvider_InjectCABundle(t *testing.T) {
	p := certproviders.SecretCertProvider{}
	err := p.InjectCABundle(nil, render.CertificateProvisionerConfig{})
	require.NoError(t, err)
}

func TestSecretCertProvider_AdditionalObjects_EmptyData(t *testing.T) {
	p := certproviders.SecretCertProvider{}
	cfg := render.CertificateProvisionerConfig{
		CertName:  "test-cert",
		Namespace: "test-ns",
	}

	objs, err := p.AdditionalObjects(cfg)
	require.NoError(t, err)
	require.Len(t, objs, 1)

	secret := objs[0]
	assert.Equal(t, "Secret", secret.GetKind())
	assert.Equal(t, "test-cert", secret.GetName())
	assert.Equal(t, "test-ns", secret.GetNamespace())

	secretType, _, _ := unstructured.NestedString(secret.Object, "type")
	assert.Equal(t, string(corev1.SecretTypeTLS), secretType)

	data, _, _ := unstructured.NestedMap(secret.Object, "data")
	assert.Contains(t, data, "tls.crt")
	assert.Contains(t, data, "tls.key")
}

func TestSecretCertProvider_AdditionalObjects_WithData(t *testing.T) {
	cert := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")
	key := []byte("-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----")

	p := certproviders.SecretCertProvider{
		Cert: cert,
		Key:  key,
	}
	cfg := render.CertificateProvisionerConfig{
		CertName:  "my-cert",
		Namespace: "my-ns",
	}

	objs, err := p.AdditionalObjects(cfg)
	require.NoError(t, err)
	require.Len(t, objs, 1)

	secret := objs[0]
	assert.Equal(t, "my-cert", secret.GetName())
	assert.Equal(t, "my-ns", secret.GetNamespace())
}

func TestSecretCertProvider_GetCertSecretInfo(t *testing.T) {
	p := certproviders.SecretCertProvider{}
	cfg := render.CertificateProvisionerConfig{
		CertName: "test-cert",
	}

	info := p.GetCertSecretInfo(cfg)
	assert.Equal(t, "test-cert", info.SecretName)
	assert.Equal(t, "tls.crt", info.CertificateKey)
	assert.Equal(t, "tls.key", info.PrivateKeyKey)
}
