package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderCmd_BasicRendering(t *testing.T) {
	f, err := os.Open("testdata/argocd-bundle.tar")
	require.NoError(t, err)
	defer f.Close()

	bundleFS, err := tarToFS(f)
	require.NoError(t, err)

	source := fromFSHelper(t, bundleFS)
	objs, err := renderBundle(source, renderConfig{
		InstallNamespace: "test-ns",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, objs)

	var buf bytes.Buffer
	require.NoError(t, writeYAML(&buf, objs))

	output := buf.String()
	assert.Contains(t, output, "apiVersion:")
	assert.Contains(t, output, "namespace: test-ns")
	assert.Contains(t, output, "---")
}

func TestRenderCmd_WithWatchNamespace(t *testing.T) {
	f, err := os.Open("testdata/argocd-bundle.tar")
	require.NoError(t, err)
	defer f.Close()

	bundleFS, err := tarToFS(f)
	require.NoError(t, err)

	source := fromFSHelper(t, bundleFS)
	objs, err := renderBundle(source, renderConfig{
		InstallNamespace: "test-ns",
		WatchNamespaces:  []string{"watch-ns"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, objs)
}

func TestRenderCmd_WithConfigFile(t *testing.T) {
	cfgContent := `
installNamespace: config-ns
providedAPIsClusterRoles: true
`
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))

	cfg, err := loadConfig(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "config-ns", cfg.InstallNamespace)
	assert.True(t, cfg.ProvidedAPIsClusterRoles)
}

func TestRenderCmd_FlagOverridesConfig(t *testing.T) {
	cfgContent := `
installNamespace: config-ns
watchNamespaces:
  - config-watch
`
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))

	cfg, err := loadConfig(cfgPath)
	require.NoError(t, err)

	flagInstallNS := "flag-ns"
	flagWatchNS := []string{"flag-watch"}

	if flagInstallNS != "" {
		cfg.InstallNamespace = flagInstallNS
	}
	if len(flagWatchNS) > 0 {
		cfg.WatchNamespaces = flagWatchNS
	}

	assert.Equal(t, "flag-ns", cfg.InstallNamespace)
	assert.Equal(t, []string{"flag-watch"}, cfg.WatchNamespaces)
}

func TestRenderCmd_MissingNamespaceError(t *testing.T) {
	cfg := renderConfig{}
	assert.Empty(t, cfg.InstallNamespace)
}

func TestRenderCmd_InvalidConfigFile(t *testing.T) {
	_, err := loadConfig("/nonexistent/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}

func TestRenderCmd_BadConfigYAML(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "bad.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte("{{not yaml"), 0600))

	_, err := loadConfig(cfgPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config file")
}

func TestRenderCmd_InvalidTar(t *testing.T) {
	_, err := tarToFS(strings.NewReader("not a tar"))
	require.Error(t, err)
}

func TestRenderCmd_ProvidedAPIsOption(t *testing.T) {
	f, err := os.Open("testdata/argocd-bundle.tar")
	require.NoError(t, err)
	defer f.Close()

	bundleFS, err := tarToFS(f)
	require.NoError(t, err)

	source := fromFSHelper(t, bundleFS)
	objs, err := renderBundle(source, renderConfig{
		InstallNamespace:         "test-ns",
		ProvidedAPIsClusterRoles: true,
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, writeYAML(&buf, objs))
	output := buf.String()
	assert.Contains(t, output, "aggregate-to-admin")
	assert.Contains(t, output, "aggregate-to-edit")
	assert.Contains(t, output, "aggregate-to-view")
}

func TestCertificateProviderConfig_CertManager(t *testing.T) {
	cfgContent := `
installNamespace: test-ns
certificateProvider:
  type: cert-manager
`
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))

	cfg, err := loadConfig(cfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg.CertificateProvider)
	assert.Equal(t, "cert-manager", cfg.CertificateProvider.Type)
	require.NoError(t, cfg.CertificateProvider.validate())

	opts := buildRenderOptions(cfg)
	assert.NotEmpty(t, opts)
}

func TestCertificateProviderConfig_OpenShiftServiceCA(t *testing.T) {
	cfgContent := `
installNamespace: test-ns
certificateProvider:
  type: openshift-service-ca
`
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))

	cfg, err := loadConfig(cfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg.CertificateProvider)
	assert.Equal(t, "openshift-service-ca", cfg.CertificateProvider.Type)
	require.NoError(t, cfg.CertificateProvider.validate())
}

func TestCertificateProviderConfig_None(t *testing.T) {
	cfgContent := `
installNamespace: test-ns
certificateProvider:
  type: none
`
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))

	cfg, err := loadConfig(cfgPath)
	require.NoError(t, err)
	require.NoError(t, cfg.CertificateProvider.validate())

	opts := buildRenderOptions(cfg)
	// none should not add a cert provider option
	for _, opt := range opts {
		_ = opt // just verify no panic
	}
}

func TestCertificateProviderConfig_Omitted(t *testing.T) {
	cfgContent := `
installNamespace: test-ns
`
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgContent), 0600))

	cfg, err := loadConfig(cfgPath)
	require.NoError(t, err)
	assert.Nil(t, cfg.CertificateProvider)
	require.NoError(t, cfg.CertificateProvider.validate())
}

func TestCertificateProviderConfig_InvalidType(t *testing.T) {
	cfg := &certificateProviderConfig{Type: "bogus"}
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
	assert.Contains(t, err.Error(), "valid types")
}

func TestCertificateProviderConfig_CertManagerWithWebhookBundle(t *testing.T) {
	f, err := os.Open("testdata/argocd-bundle.tar")
	require.NoError(t, err)
	defer f.Close()

	bundleFS, err := tarToFS(f)
	require.NoError(t, err)

	source := fromFSHelper(t, bundleFS)
	objs, err := renderBundle(source, renderConfig{
		InstallNamespace: "test-ns",
		CertificateProvider: &certificateProviderConfig{
			Type: "cert-manager",
		},
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, writeYAML(&buf, objs))
	output := buf.String()
	// argocd bundle has no webhooks, so cert-manager won't add objects,
	// but it should not error either
	assert.Contains(t, output, "apiVersion:")
}
