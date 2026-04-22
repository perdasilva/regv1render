package regv1render

import (
	"github.com/perdasilva/regv1render/internal/render/certproviders"
)

// CertManagerProvider is a CertificateProvider that uses cert-manager
// to provision TLS certificates for webhooks.
type CertManagerProvider = certproviders.CertManagerCertificateProvider

// OpenShiftServiceCAProvider is a CertificateProvider that uses the
// OpenShift service-ca operator to provision TLS certificates.
type OpenShiftServiceCAProvider = certproviders.OpenshiftServiceCaCertificateProvider
