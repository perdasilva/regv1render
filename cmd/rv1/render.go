package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing/fstest"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/perdasilva/regv1render"
)

type renderConfig struct {
	ProvidedAPIsClusterRoles bool                          `json:"providedAPIsClusterRoles"`
	DeploymentConfig         *regv1render.DeploymentConfig `json:"deploymentConfig,omitempty"`
	CertificateProvider      *certificateProviderConfig    `json:"certificateProvider,omitempty"`
}

const (
	certProviderNone               = "none"
	certProviderCertManager        = "cert-manager"
	certProviderOpenShiftServiceCA = "openshift-service-ca"
)

type certificateProviderConfig struct {
	Type string `json:"type"`
}

var validCertProviderTypes = []string{certProviderNone, certProviderCertManager, certProviderOpenShiftServiceCA}

func (c *certificateProviderConfig) validate() error {
	if c == nil || c.Type == "" || c.Type == certProviderNone {
		return nil
	}
	if !slices.Contains(validCertProviderTypes, c.Type) {
		return fmt.Errorf("unknown certificate provider type %q (valid types: %v)", c.Type, validCertProviderTypes)
	}
	return nil
}

func renderCmd() *cobra.Command {
	var (
		installNamespace string
		watchNamespaces  []string
		configFile       string
	)

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render a registry+v1 bundle to plain Kubernetes manifests",
		Long: `Reads a registry+v1 bundle as a tar stream from stdin,
renders it to plain Kubernetes manifests, and writes
multi-document YAML to stdout.

Namespace modes (controlled by --watch-namespace):
  AllNamespaces    Omit the flag (default when bundle supports it)
  OwnNamespace    Set to the same value as --install-namespace
  SingleNamespace  Set to a different namespace
  MultiNamespace   Repeat the flag for each namespace

The --config flag accepts a YAML file with renderer options:
  providedAPIsClusterRoles, deploymentConfig, and
  certificateProvider (type: cert-manager, openshift-service-ca,
  or none).

Examples:
  # Render with AllNamespaces (default)
  crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns

  # Render with SingleNamespace
  crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns --watch-namespace target-ns

  # Render with MultiNamespace
  crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns --watch-namespace ns1 --watch-namespace ns2

  # Render with a config file
  crane export quay.io/my/bundle:v1 - | rv1 render --config render.yaml`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if installNamespace == "" {
				return fmt.Errorf("--install-namespace is required")
			}

			cfg, err := loadConfig(configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err := cfg.CertificateProvider.validate(); err != nil {
				return fmt.Errorf("invalid config: %w", err)
			}

			bundleFS, err := tarToFS(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading bundle from stdin: %w", err)
			}

			source := regv1render.FromFS(bundleFS)
			rv1, err := source.GetBundle()
			if err != nil {
				return fmt.Errorf("parsing bundle: %w", err)
			}

			renderer := buildRenderer(cfg)
			renderOpts := buildRenderOptions(cfg, watchNamespaces)
			objs, err := renderer.Render(rv1, installNamespace, renderOpts...)
			if err != nil {
				return fmt.Errorf("rendering bundle: %w", err)
			}

			return writeYAML(os.Stdout, objs)
		},
	}

	cmd.Flags().StringVar(&installNamespace, "install-namespace", "", "namespace where the operator will be installed (required)")
	cmd.Flags().StringSliceVar(&watchNamespaces, "watch-namespace", nil, "namespace(s) the operator should watch (repeatable)")
	cmd.Flags().StringVar(&configFile, "config", "", "path to YAML config file with rendering options")

	return cmd
}

func loadConfig(path string) (renderConfig, error) {
	if path == "" {
		return renderConfig{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return renderConfig{}, fmt.Errorf("reading config file %q: %w", path, err)
	}
	var cfg renderConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return renderConfig{}, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	return cfg, nil
}

func buildRenderer(cfg renderConfig) *regv1render.Renderer {
	b := regv1render.NewRendererBuilder()
	if cfg.DeploymentConfig != nil {
		b.WithDeploymentConfig(cfg.DeploymentConfig)
	}
	if p := cfg.CertificateProvider; p != nil {
		switch p.Type {
		case certProviderCertManager:
			b.WithCertificateProvider(regv1render.CertManagerProvider{})
		case certProviderOpenShiftServiceCA:
			b.WithCertificateProvider(regv1render.OpenShiftServiceCAProvider{})
		}
	}
	return b.Build()
}

func buildRenderOptions(cfg renderConfig, watchNamespaces []string) []regv1render.RenderOption {
	var opts []regv1render.RenderOption
	if len(watchNamespaces) > 0 {
		opts = append(opts, regv1render.WithTargetNamespaces(watchNamespaces...))
	}
	if cfg.ProvidedAPIsClusterRoles {
		opts = append(opts, regv1render.WithProvidedAPIsClusterRoles())
	}
	return opts
}

func writeYAML(w io.Writer, objs []client.Object) error {
	for i, obj := range objs {
		if i > 0 {
			if _, err := fmt.Fprintln(w, "---"); err != nil {
				return err
			}
		}
		data, err := yaml.Marshal(obj)
		if err != nil {
			return fmt.Errorf("marshaling %s %q: %w", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName(), err)
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func tarToFS(r io.Reader) (fs.FS, error) {
	mapFS := make(fstest.MapFS)
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar entry: %w", err)
		}

		name := filepath.Clean(header.Name)
		name = strings.TrimPrefix(name, "/")
		name = strings.TrimPrefix(name, "./")

		if header.Typeflag == tar.TypeDir {
			continue
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("reading tar entry %q: %w", name, err)
		}

		mapFS[name] = &fstest.MapFile{
			Data: data,
			Mode: fs.FileMode(header.Mode),
		}
	}

	return mapFS, nil
}
