package certproviders

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateProvider encapsulates the creation and modification of objects for certificate provisioning
// in Kubernetes by vendors such as CertManager or the OpenshiftServiceCA operator.
type CertificateProvider interface {
	InjectCABundle(obj client.Object, cfg CertificateProvisionerConfig) error
	AdditionalObjects(cfg CertificateProvisionerConfig) ([]unstructured.Unstructured, error)
	GetCertSecretInfo(cfg CertificateProvisionerConfig) CertSecretInfo
}

// CertSecretInfo describes the certificate secret resource information.
type CertSecretInfo struct {
	SecretName     string
	CertificateKey string
	PrivateKeyKey  string
}

// CertificateProvisionerConfig contains the necessary information for a CertificateProvider
// to correctly generate and modify objects for certificate injection and automation.
type CertificateProvisionerConfig struct {
	ServiceName  string
	CertName     string
	Namespace    string
	CertProvider CertificateProvider
}

// CertificateProvisioner uses a CertificateProvider to modify and generate objects based on its
// CertificateProvisionerConfig.
type CertificateProvisioner CertificateProvisionerConfig

func (c CertificateProvisioner) InjectCABundle(obj client.Object) error {
	if c.CertProvider == nil {
		return nil
	}
	return c.CertProvider.InjectCABundle(obj, CertificateProvisionerConfig(c))
}

func (c CertificateProvisioner) AdditionalObjects() ([]unstructured.Unstructured, error) {
	if c.CertProvider == nil {
		return nil, nil
	}
	return c.CertProvider.AdditionalObjects(CertificateProvisionerConfig(c))
}

func (c CertificateProvisioner) GetCertSecretInfo() *CertSecretInfo {
	if c.CertProvider == nil {
		return nil
	}
	info := c.CertProvider.GetCertSecretInfo(CertificateProvisionerConfig(c))
	return &info
}
