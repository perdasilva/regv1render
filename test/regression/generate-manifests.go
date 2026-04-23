package main

import (
	"cmp"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/perdasilva/rv1/internal/bundle/source"
	"github.com/perdasilva/rv1/internal/render"
	"github.com/perdasilva/rv1/internal/render/certproviders"
)

func main() {
	bundleRootDir := "testdata/bundles/"
	defaultOutputDir := "./testdata/tmp/rendered/"
	outputRootDir := flag.String("output-dir", defaultOutputDir, "path to write rendered manifests to")
	flag.Parse()

	if err := os.RemoveAll(*outputRootDir); err != nil {
		fmt.Printf("error removing output directory: %v\n", err)
		os.Exit(1)
	}

	for _, tc := range []struct {
		name             string
		installNamespace string
		watchNamespace   string
		bundle           string
		testCaseName     string
		configureBuilder func(*render.RendererBuilder)
		renderOpts       []render.RenderOption
	}{
		{
			name:             "AllNamespaces",
			installNamespace: "argocd-system",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "all-namespaces",
		}, {
			name:             "SingleNamespaces",
			installNamespace: "argocd-system",
			watchNamespace:   "argocd-watch",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "single-namespace",
		}, {
			name:             "OwnNamespaces",
			installNamespace: "argocd-system",
			watchNamespace:   "argocd-system",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "own-namespace",
		}, {
			name:             "Webhooks",
			installNamespace: "webhook-system",
			bundle:           "webhook-operator.v0.0.5",
			testCaseName:     "all-webhook-types",
		}, {
			name:             "WithDeploymentConfigOptions",
			installNamespace: "argocd-system",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "with-deploymentconfig-options",
			configureBuilder: func(b *render.RendererBuilder) {
				b.WithDeploymentConfig(&render.DeploymentConfig{
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "some/key",
							Operator: corev1.TolerationOpEqual,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:               "someother/key",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: ptr.To(int64(120)),
						},
					},
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "CUSTOM_ENV_VAR",
							Value: "custom-value",
						},
						{
							Name:  "LOG_LEVEL",
							Value: "debug",
						},
					},
					EnvFrom: []corev1.EnvFromSource{
						{
							Prefix: "test",
						},
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "configmapForTest",
								},
							},
						},
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secretForTest",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "test-configmap-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "testVolumeConfigMap",
									},
								},
							},
						},
						{
							Name:         "test-emptydir-volume",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "test-configmap-volume",
							MountPath: "/test-volume-mount",
							ReadOnly:  true,
						},
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"linux"},
											},
										},
										MatchFields: []corev1.NodeSelectorRequirement{
											{
												Key:      "key",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"val1", "val2"},
											},
										},
									},
								},
							},
						},
						PodAffinity: &corev1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"app.kubernetes.io/name": "test-app",
										},
									},
									TopologyKey: "topKey",
									Namespaces:  []string{"test", "test2"},
								},
							},
						},
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"app.kubernetes.io/name": "test-app-2",
											},
										},
										TopologyKey: "topKey2",
									},
								},
							},
						},
					},
					Annotations: map[string]string{
						"foo":     "bar",
						"testkey": "testval",
					},
				})
			},
		}, {
			name:             "WithEmptyAffinityConfig",
			installNamespace: "argocd-system",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "with-empty-affinity",
			configureBuilder: func(b *render.RendererBuilder) {
				b.WithDeploymentConfig(&render.DeploymentConfig{
					Affinity: &corev1.Affinity{},
				})
			},
		}, {
			name:             "WithEmptyAffinitySubTypeConfig",
			installNamespace: "argocd-system",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "with-empty-affinity-subtype",
			configureBuilder: func(b *render.RendererBuilder) {
				b.WithDeploymentConfig(&render.DeploymentConfig{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{},
					},
				})
			},
		}, {
			name:             "WithProvidedAPIsClusterRoles",
			installNamespace: "argocd-system",
			bundle:           "argocd-operator.v0.6.0",
			testCaseName:     "with-provided-apis-clusterroles",
			renderOpts: []render.RenderOption{
				render.WithProvidedAPIsClusterRoles(),
			},
		}, {
			name:             "WithSecretCertProvider",
			installNamespace: "webhook-system",
			bundle:           "webhook-operator.v0.0.5",
			testCaseName:     "with-secret-cert-provider",
			configureBuilder: func(b *render.RendererBuilder) {
				b.WithCertificateProvider(certproviders.SecretCertProvider{})
			},
		},
	} {
		bundlePath := filepath.Join(bundleRootDir, tc.bundle)
		generatedManifestPath := filepath.Join(*outputRootDir, tc.bundle, tc.testCaseName)
		if err := generateManifests(generatedManifestPath, bundlePath, tc.installNamespace, tc.watchNamespace, tc.configureBuilder, tc.renderOpts); err != nil {
			fmt.Printf("Error generating manifests: %v", err)
			os.Exit(1)
		}
	}
}

func generateManifests(outputPath, bundleDir, installNamespace, watchNamespace string, configureBuilder func(*render.RendererBuilder), extraRenderOpts []render.RenderOption) error {
	regv1, err := source.FromFS(os.DirFS(bundleDir)).GetBundle()
	if err != nil {
		fmt.Printf("error parsing bundle directory: %v\n", err)
		os.Exit(1)
	}

	b := render.NewRendererBuilder()
	if configureBuilder != nil {
		configureBuilder(b)
	}
	renderer := b.Build()

	var renderOpts []render.RenderOption
	if watchNamespace != "" {
		renderOpts = append(renderOpts, render.WithTargetNamespaces(watchNamespace))
	}
	renderOpts = append(renderOpts, extraRenderOpts...)
	objs, err := renderer.Render(regv1, installNamespace, renderOpts...)
	if err != nil {
		return fmt.Errorf("error converting registry+v1 bundle: %w", err)
	}

	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating bundle directory: %w", err)
	}

	if err := func() error {
		for idx, obj := range slices.SortedFunc(slices.Values(objs), orderByKindNamespaceName) {
			kind := obj.GetObjectKind().GroupVersionKind().Kind
			fileName := fmt.Sprintf("%02d_%s_%s.yaml", idx, strings.ToLower(kind), obj.GetName())
			data, err := yaml.Marshal(obj)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(outputPath, fileName), data, 0600); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		_ = os.RemoveAll(outputPath)
		return fmt.Errorf("error writing object files: %w", err)
	}

	return nil
}

func orderByKindNamespaceName(a client.Object, b client.Object) int {
	return cmp.Or(
		cmp.Compare(a.GetObjectKind().GroupVersionKind().Kind, b.GetObjectKind().GroupVersionKind().Kind),
		cmp.Compare(a.GetNamespace(), b.GetNamespace()),
		cmp.Compare(a.GetName(), b.GetName()),
	)
}
