package render

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/perdasilva/rv1/internal/render/certproviders"
)

func Test_certProvisionerFor(t *testing.T) {
	mockProvider := certproviders.NewMockCertificateProvider(t)
	prov := certProvisionerFor("my.deployment.thing", options{
		InstallNamespace:    "my-namespace",
		CertificateProvider: mockProvider,
	})

	require.Equal(t, prov.CertProvider, mockProvider)
	require.Equal(t, "my-deployment-thing-service", prov.ServiceName)
	require.Equal(t, "my-deployment-thing-service-cert", prov.CertName)
	require.Equal(t, "my-namespace", prov.Namespace)
}

func Test_certProvisionerFor_ExtraLargeName_MoreThan63Chars(t *testing.T) {
	prov := certProvisionerFor("my.object.thing.has.a.really.really.really.really.really.long.name", options{})

	require.Len(t, prov.ServiceName, 63)
	require.Len(t, prov.CertName, 63)
	require.Equal(t, "my-object-thing-has-a-really-really-really-really-reall-service", prov.ServiceName)
	require.Equal(t, "my-object-thing-has-a-really-really-really-really-reall-se-cert", prov.CertName)
}
