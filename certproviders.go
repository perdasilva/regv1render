package regv1render

import (
	"github.com/perdasilva/regv1render/internal/render/certproviders"
)

type certManagerProvider = certproviders.CertManagerCertificateProvider
type openShiftServiceCAProvider = certproviders.OpenshiftServiceCaCertificateProvider
