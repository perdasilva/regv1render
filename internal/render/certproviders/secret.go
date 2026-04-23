package certproviders

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/rv1/internal/render"
)

var _ render.CertificateProvider = (*SecretCertProvider)(nil)

// SecretCertProvider generates a kubernetes.io/tls Secret for
// webhook TLS. If Cert and Key are empty, the Secret is created with
// empty data so users can populate it externally.
type SecretCertProvider struct {
	Cert []byte
	Key  []byte
}

func (p SecretCertProvider) InjectCABundle(_ client.Object, _ render.CertificateProvisionerConfig) error {
	return nil
}

func (p SecretCertProvider) AdditionalObjects(cfg render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.CertName,
			Namespace: cfg.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       p.Cert,
			corev1.TLSPrivateKeyKey: p.Key,
		},
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return nil, err
	}

	u := unstructured.Unstructured{Object: obj}
	u.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
	return []unstructured.Unstructured{u}, nil
}

func (p SecretCertProvider) GetCertSecretInfo(cfg render.CertificateProvisionerConfig) render.CertSecretInfo {
	return render.CertSecretInfo{
		SecretName:     cfg.CertName,
		CertificateKey: corev1.TLSCertKey,
		PrivateKeyKey:  corev1.TLSPrivateKeyKey,
	}
}
