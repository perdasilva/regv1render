package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/perdasilva/regv1render"
)

type renderConfig struct {
	InstallNamespace         string                        `json:"installNamespace"`
	WatchNamespaces          []string                      `json:"watchNamespaces"`
	ProvidedAPIsClusterRoles bool                          `json:"providedAPIsClusterRoles"`
	DeploymentConfig         *regv1render.DeploymentConfig `json:"deploymentConfig,omitempty"`
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

Examples:
  # Render from a container image using crane
  crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns

  # Render with watch namespaces
  crane export quay.io/my/bundle:v1 - | rv1 render --install-namespace my-ns --watch-namespace ns1 --watch-namespace ns2

  # Render with a config file
  crane export quay.io/my/bundle:v1 - | rv1 render --config render.yaml`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if installNamespace != "" {
				cfg.InstallNamespace = installNamespace
			}
			if len(watchNamespaces) > 0 {
				cfg.WatchNamespaces = watchNamespaces
			}

			if cfg.InstallNamespace == "" {
				return fmt.Errorf("--install-namespace is required (or set installNamespace in config file)")
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

			opts := buildRenderOptions(cfg)
			objs, err := regv1render.Render(rv1, cfg.InstallNamespace, opts...)
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

func buildRenderOptions(cfg renderConfig) []regv1render.Option {
	var opts []regv1render.Option
	if len(cfg.WatchNamespaces) > 0 {
		opts = append(opts, regv1render.WithTargetNamespaces(cfg.WatchNamespaces...))
	}
	if cfg.ProvidedAPIsClusterRoles {
		opts = append(opts, regv1render.WithProvidedAPIsClusterRoles())
	}
	if cfg.DeploymentConfig != nil {
		opts = append(opts, regv1render.WithDeploymentConfig(cfg.DeploymentConfig))
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
