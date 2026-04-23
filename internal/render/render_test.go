package render_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/regv1render/internal/bundle"
	"github.com/perdasilva/regv1render/internal/render"
	"github.com/perdasilva/regv1render/internal/util/testutil/clusterserviceversion"
)

type failingValidator struct{ err error }

func (f failingValidator) Validate(_ *bundle.RegistryV1) error { return f.err }

func Test_Renderer_NoConfig(t *testing.T) {
	renderer := render.NewRendererBuilder(nil, nil).Build()
	objs, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "")
	require.NoError(t, err)
	require.Empty(t, objs)
}

func Test_Renderer_ValidatesBundle(t *testing.T) {
	renderer := render.NewRendererBuilder(failingValidator{err: errors.New("this bundle is invalid")}, nil).Build()
	objs, err := renderer.Render(bundle.RegistryV1{}, "")
	require.Nil(t, objs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "this bundle is invalid")
}

func Test_Renderer_DefaultTargetNamespaces(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		supportedInstallModes []v1alpha1.InstallModeType
		expectedErrMsg        string
	}{
		{
			name:                  "Default to AllNamespaces when only AllNamespaces",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeAllNamespaces},
		},
		{
			name:                  "Default to AllNamespaces when AllNamespaces + OwnNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeAllNamespaces, v1alpha1.InstallModeTypeOwnNamespace},
		},
		{
			name:                  "Default to AllNamespaces when AllNamespaces + SingleNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeAllNamespaces, v1alpha1.InstallModeTypeSingleNamespace},
		},
		{
			name:                  "Default to AllNamespaces when AllNamespaces + MultiNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeAllNamespaces, v1alpha1.InstallModeTypeMultiNamespace},
		},
		{
			name:                  "Default to AllNamespaces when all modes supported",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeAllNamespaces, v1alpha1.InstallModeTypeOwnNamespace, v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeMultiNamespace},
		},
		{
			name:                  "No default when only OwnNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeOwnNamespace},
			expectedErrMsg:        "exactly one target namespace must be specified",
		},
		{
			name:                  "No default when only SingleNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeSingleNamespace},
			expectedErrMsg:        "exactly one target namespace must be specified",
		},
		{
			name:                  "No default when SingleNamespace + OwnNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeOwnNamespace},
			expectedErrMsg:        "exactly one target namespace must be specified",
		},
		{
			name:                  "No default when only MultiNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeMultiNamespace},
			expectedErrMsg:        "at least one target namespace must be specified",
		},
		{
			name:                  "No default when SingleNamespace + MultiNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeMultiNamespace},
			expectedErrMsg:        "at least one target namespace must be specified",
		},
		{
			name:                  "No default when OwnNamespace + MultiNamespace",
			supportedInstallModes: []v1alpha1.InstallModeType{v1alpha1.InstallModeTypeOwnNamespace, v1alpha1.InstallModeTypeMultiNamespace},
			expectedErrMsg:        "at least one target namespace must be specified",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			renderer := render.NewRendererBuilder(nil, nil).Build()
			_, err := renderer.Render(bundle.RegistryV1{
				CSV: clusterserviceversion.Builder().
					WithName("test").
					WithInstallModeSupportFor(tc.supportedInstallModes...).Build(),
			}, "some-namespace")
			if tc.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_Renderer_ValidatesTargetNamespaces(t *testing.T) {
	for _, tc := range []struct {
		name             string
		installNamespace string
		csv              v1alpha1.ClusterServiceVersion
		targetNamespaces []string
		errMsg           string
	}{
		{
			name:             "accepts empty targetNamespaces (uses default)",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
			targetNamespaces: []string{},
		},
		{
			name:             "rejects all namespace if not supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace).Build(),
			targetNamespaces: []string{""},
			errMsg:           "invalid target namespaces []: supported install modes [SingleNamespace] do not support targeting all namespaces",
		},
		{
			name:             "rejects own namespace if only AllNamespaces supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
			targetNamespaces: []string{"install-namespace"},
			errMsg:           "invalid target namespaces [install-namespace]: supported install modes [AllNamespaces] do not support targeting own namespace",
		},
		{
			name:             "rejects out of own namespace if only OwnNamespace supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).Build(),
			targetNamespaces: []string{"not-install-namespace"},
			errMsg:           "invalid target namespaces [not-install-namespace]: supported install modes [OwnNamespace] do not support target namespaces [not-install-namespace]",
		},
		{
			name:             "rejects multi-namespace if not supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
			targetNamespaces: []string{"ns1", "ns2", "ns3"},
			errMsg:           "invalid target namespaces [ns1 ns2 ns3]: supported install modes [AllNamespaces] do not support target namespaces [ns1 ns2 ns3]",
		},
		{
			name:             "rejects if bundle supports no install modes",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().Build(),
			targetNamespaces: []string{"some-namespace"},
			errMsg:           "invalid target namespaces [some-namespace]: supported install modes [] do not support target namespaces [some-namespace]",
		},
		{
			name:             "rejects multi with own namespace if OwnNamespace not supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).Build(),
			targetNamespaces: []string{"n1", "n2", "install-namespace"},
			errMsg:           "invalid target namespaces [n1 n2 install-namespace]: supported install modes [MultiNamespace] do not support targeting own namespace",
		},
		{
			name:             "accepts all namespace when supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
			targetNamespaces: []string{""},
		},
		{
			name:             "accepts own namespace when supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).Build(),
			targetNamespaces: []string{"install-namespace"},
		},
		{
			name:             "accepts single namespace when supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace).Build(),
			targetNamespaces: []string{"some-namespace"},
		},
		{
			name:             "accepts multi namespace when supported",
			installNamespace: "install-namespace",
			csv:              clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).Build(),
			targetNamespaces: []string{"n1", "n2", "n3"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			renderer := render.NewRendererBuilder(nil, nil).Build()
			_, err := renderer.Render(
				bundle.RegistryV1{CSV: tc.csv},
				tc.installNamespace,
				render.WithTargetNamespaces(tc.targetNamespaces...),
			)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.errMsg, err.Error())
			}
		})
	}
}

func Test_Renderer_WithDeploymentConfig(t *testing.T) {
	expectedConfig := &render.DeploymentConfig{
		Env: []corev1.EnvVar{
			{Name: "TEST_ENV", Value: "test-value"},
		},
	}

	var receivedConfig *render.DeploymentConfig
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			receivedConfig = opts.DeploymentConfig
			return nil, nil
		},
	}).WithDeploymentConfig(expectedConfig).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		},
		"test-namespace",
	)

	require.NoError(t, err)
	require.Equal(t, expectedConfig, receivedConfig)
}

func Test_Renderer_DeploymentConfig_NilWhenNotProvided(t *testing.T) {
	var receivedConfig *render.DeploymentConfig
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			receivedConfig = opts.DeploymentConfig
			return nil, nil
		},
	}).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
	require.Nil(t, receivedConfig)
}

func Test_Renderer_CallsGeneratorsAndAggregatesOutput(t *testing.T) {
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			return []client.Object{&corev1.Namespace{}, &corev1.Service{}}, nil
		},
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			return []client.Object{&corev1.ConfigMap{}}, nil
		},
	}).Build()
	objs, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "")
	require.NoError(t, err)
	require.Len(t, objs, 3)
}

func Test_Renderer_ReturnsGeneratorErrors(t *testing.T) {
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			return []client.Object{&corev1.Namespace{}}, nil
		},
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			return nil, errors.New("generator error")
		},
	}).Build()
	objs, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "")
	require.Nil(t, objs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "generator error")
}

func Test_Renderer_WithCertificateProvider(t *testing.T) {
	var receivedProvider render.CertificateProvider
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			receivedProvider = opts.CertificateProvider
			return nil, nil
		},
	}).WithCertificateProvider(fakeCertProvider{}).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
	require.NotNil(t, receivedProvider)
}

func Test_Renderer_WithUniqueNameGenerator(t *testing.T) {
	var receivedName string
	customGen := func(base string, obj interface{}) string { return "custom-name" }
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			receivedName = opts.UniqueNameGenerator("base", nil)
			return nil, nil
		},
	}).WithUniqueNameGenerator(customGen).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
	require.Equal(t, "custom-name", receivedName)
}

func Test_Renderer_DefaultOptionsPassedToGenerators(t *testing.T) {
	var received render.Options
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			received = opts
			return nil, nil
		},
	}).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "install-ns")
	require.NoError(t, err)
	require.Equal(t, "install-ns", received.InstallNamespace)
	require.Equal(t, []string{""}, received.TargetNamespaces)
	require.NotNil(t, received.UniqueNameGenerator)
}

func Test_Renderer_DeploymentConfig_NilWhenExplicitlyNil(t *testing.T) {
	var receivedConfig *render.DeploymentConfig
	renderer := render.NewRendererBuilder(nil, []render.ResourceGenerator{
		func(rv1 *bundle.RegistryV1, opts render.Options) ([]client.Object, error) {
			receivedConfig = opts.DeploymentConfig
			return nil, nil
		},
	}).WithDeploymentConfig(nil).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			CSV: clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
	require.Nil(t, receivedConfig)
}

type fakeCertProvider struct{}

func (f fakeCertProvider) InjectCABundle(_ client.Object, _ render.CertificateProvisionerConfig) error {
	return nil
}
func (f fakeCertProvider) AdditionalObjects(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
	return nil, nil
}
func (f fakeCertProvider) GetCertSecretInfo(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
	return render.CertSecretInfo{}
}
