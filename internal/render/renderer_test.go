package render_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/perdasilva/rv1/internal/bundle"
	"github.com/perdasilva/rv1/internal/render"
	. "github.com/perdasilva/rv1/internal/util/testutil"
	"github.com/perdasilva/rv1/internal/util/testutil/clusterserviceversion"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func findObjectsByKind[T client.Object](objs []client.Object) []T {
	var result []T
	for _, obj := range objs {
		if typed, ok := obj.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

func findObjectByKindAndName[T client.Object](objs []client.Object, name string) T {
	for _, obj := range objs {
		if typed, ok := obj.(T); ok {
			if typed.GetName() == name {
				return typed
			}
		}
	}
	var zero T
	return zero
}

// fakeNameGen returns a predictable unique name generator for tests.
func fakeNameGen() render.UniqueNameGenerator {
	return func(base string, _ interface{}) string {
		return base
	}
}

// buildRenderer creates a *render.Renderer with no validator, a fake name generator,
// and optional builder configuration.
func buildRenderer(opts ...func(*render.RendererBuilder)) *render.Renderer {
	b := render.NewRendererBuilder().WithUniqueNameGenerator(fakeNameGen())
	for _, opt := range opts {
		opt(b)
	}
	return b.Build()
}

// withCertProvider configures a FakeCertProvider on the builder.
func withCertProvider(cp render.CertificateProvider) func(*render.RendererBuilder) {
	return func(b *render.RendererBuilder) {
		b.WithCertificateProvider(cp)
	}
}

// withDeploymentConfig configures a DeploymentConfig on the builder.
func withDeploymentConfig(dc *render.DeploymentConfig) func(*render.RendererBuilder) {
	return func(b *render.RendererBuilder) {
		b.WithDeploymentConfig(dc)
	}
}

// ---------------------------------------------------------------------------
// Deployment Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_Deployment_GeneratesDeploymentResources(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithAnnotations(map[string]string{
				"csv": "annotation",
			}).
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{
					Name: "deployment-one",
					Label: map[string]string{
						"bar": "foo",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"pod": "annotation",
								},
							},
							Spec: corev1.PodSpec{
								ServiceAccountName: "some-service-account",
							},
						},
					},
				},
				v1alpha1.StrategyDeploymentSpec{
					Name: "deployment-two",
					Spec: appsv1.DeploymentSpec{},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	deps := findObjectsByKind[*appsv1.Deployment](objs)
	require.Len(t, deps, 2)

	// Sort by name for stable assertion order
	slices.SortFunc(deps, func(a, b *appsv1.Deployment) int {
		return cmp.Compare(a.Name, b.Name)
	})

	depOne := deps[0]
	require.Equal(t, "deployment-one", depOne.Name)
	require.Equal(t, "install-namespace", depOne.Namespace)
	require.Equal(t, map[string]string{"bar": "foo"}, depOne.Labels)
	require.Equal(t, ptr.To(int32(1)), depOne.Spec.RevisionHistoryLimit)
	require.Equal(t, "watch-namespace-one,watch-namespace-two", depOne.Spec.Template.Annotations["olm.targetNamespaces"])
	require.Equal(t, "annotation", depOne.Spec.Template.Annotations["csv"])
	require.Equal(t, "annotation", depOne.Spec.Template.Annotations["pod"])
	require.Equal(t, "some-service-account", depOne.Spec.Template.Spec.ServiceAccountName)

	depTwo := deps[1]
	require.Equal(t, "deployment-two", depTwo.Name)
	require.Equal(t, "install-namespace", depTwo.Namespace)
	require.Equal(t, ptr.To(int32(1)), depTwo.Spec.RevisionHistoryLimit)
	require.Equal(t, "watch-namespace-one,watch-namespace-two", depTwo.Spec.Template.Annotations["olm.targetNamespaces"])
	require.Equal(t, "annotation", depTwo.Spec.Template.Annotations["csv"])
}

func Test_Generators_Deployment_WithCertProvider(t *testing.T) {
	fakeProvider := render.FakeCertProvider{
		GetCertSecretInfoFn: func(cfg render.CertificateProvisionerConfig) render.CertSecretInfo {
			return render.CertSecretInfo{
				SecretName:     "some-secret",
				CertificateKey: "some-cert-key",
				PrivateKeyKey:  "some-private-key-key",
			}
		},
		InjectCABundleFn: func(_ client.Object, _ render.CertificateProvisionerConfig) error {
			return nil
		},
		AdditionalObjectsFn: func(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return nil, nil
		},
	}

	renderer := buildRenderer(withCertProvider(fakeProvider))

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ValidatingAdmissionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "deployment-one",
				}).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{
					Name: "deployment-one",
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "some-other-mount",
										VolumeSource: corev1.VolumeSource{
											EmptyDir: &corev1.EmptyDirVolumeSource{},
										},
									},
									{
										Name: "webhook-cert",
										VolumeSource: corev1.VolumeSource{
											EmptyDir: &corev1.EmptyDirVolumeSource{},
										},
									},
									{
										Name: "some-mount",
										VolumeSource: corev1.VolumeSource{
											EmptyDir: &corev1.EmptyDirVolumeSource{},
										},
									},
									{
										Name: "some-webhook-cert-mount",
										VolumeSource: corev1.VolumeSource{
											EmptyDir: &corev1.EmptyDirVolumeSource{},
										},
									},
								},
								Containers: []corev1.Container{
									{
										Name: "container-1",
										VolumeMounts: []corev1.VolumeMount{
											{Name: "webhook-cert", MountPath: "/webhook-cert-path"},
											{Name: "some-other-mount", MountPath: "/some/other/mount/path"},
											{Name: "some-webhook-cert-mount", MountPath: "/tmp/k8s-webhook-server/serving-certs"},
											{Name: "some-mount", MountPath: "/apiserver.local.config/certificates"},
										},
									},
									{
										Name: "container-2",
									},
								},
							},
						},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	dep := findObjectByKindAndName[*appsv1.Deployment](objs, "deployment-one")
	require.NotNil(t, dep)

	// Verify volumes
	require.Equal(t, []corev1.Volume{
		{
			Name: "some-other-mount",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "webhook-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "some-secret",
					Items: []corev1.KeyToPath{
						{Key: "some-cert-key", Path: "tls.crt"},
						{Key: "some-private-key-key", Path: "tls.key"},
					},
				},
			},
		},
		{
			Name: "apiservice-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "some-secret",
					Items: []corev1.KeyToPath{
						{Key: "some-cert-key", Path: "apiserver.crt"},
						{Key: "some-private-key-key", Path: "apiserver.key"},
					},
				},
			},
		},
	}, dep.Spec.Template.Spec.Volumes)

	// Container-1: protected mounts removed, cert mounts appended
	require.Equal(t, []corev1.VolumeMount{
		{Name: "some-other-mount", MountPath: "/some/other/mount/path"},
		{Name: "webhook-cert", MountPath: "/tmp/k8s-webhook-server/serving-certs"},
		{Name: "apiservice-cert", MountPath: "/apiserver.local.config/certificates"},
	}, dep.Spec.Template.Spec.Containers[0].VolumeMounts)

	// Container-2: cert mounts injected
	require.Equal(t, []corev1.VolumeMount{
		{Name: "webhook-cert", MountPath: "/tmp/k8s-webhook-server/serving-certs"},
		{Name: "apiservice-cert", MountPath: "/apiserver.local.config/certificates"},
	}, dep.Spec.Template.Spec.Containers[1].VolumeMounts)
}

func Test_Generators_Deployment_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			Build(),
	}, "install-namespace")
	require.NoError(t, err)
	deps := findObjectsByKind[*appsv1.Deployment](objs)
	require.Empty(t, deps)
}

// ---------------------------------------------------------------------------
// Permissions Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_Permissions_NoResourcesInAllNamespacesMode(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-one",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces(""))
	require.NoError(t, err)

	roles := findObjectsByKind[*rbacv1.Role](objs)
	require.Empty(t, roles)
	roleBindings := findObjectsByKind[*rbacv1.RoleBinding](objs)
	require.Empty(t, roleBindings)
}

func Test_Generators_Permissions_GeneratesRoleAndRoleBinding(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeOwnNamespace).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-one",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
						{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces("watch-namespace"))
	require.NoError(t, err)

	role := findObjectByKindAndName[*rbacv1.Role](objs, "csv-service-account-one")
	require.NotNil(t, role)
	require.Equal(t, "watch-namespace", role.Namespace)
	require.Equal(t, []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
		{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
	}, role.Rules)

	rb := findObjectByKindAndName[*rbacv1.RoleBinding](objs, "csv-service-account-one")
	require.NotNil(t, rb)
	require.Equal(t, "watch-namespace", rb.Namespace)
	require.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: "csv-service-account-one"}, rb.RoleRef)
	require.Equal(t, []rbacv1.Subject{{Kind: "ServiceAccount", Name: "service-account-one", Namespace: "install-namespace"}}, rb.Subjects)
}

func Test_Generators_Permissions_PerTargetNamespace(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-one",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace", "watch-namespace-two"),
	)
	require.NoError(t, err)

	roles := findObjectsByKind[*rbacv1.Role](objs)
	require.Len(t, roles, 2)

	namespaces := []string{roles[0].Namespace, roles[1].Namespace}
	slices.Sort(namespaces)
	require.Equal(t, []string{"watch-namespace", "watch-namespace-two"}, namespaces)

	rbs := findObjectsByKind[*rbacv1.RoleBinding](objs)
	require.Len(t, rbs, 2)
}

func Test_Generators_Permissions_MultiplePermissions(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeOwnNamespace).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-one",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-two",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces("watch-namespace"))
	require.NoError(t, err)

	roles := findObjectsByKind[*rbacv1.Role](objs)
	require.Len(t, roles, 2)

	r1 := findObjectByKindAndName[*rbacv1.Role](objs, "csv-service-account-one")
	require.NotNil(t, r1)
	r2 := findObjectByKindAndName[*rbacv1.Role](objs, "csv-service-account-two")
	require.NotNil(t, r2)

	rbs := findObjectsByKind[*rbacv1.RoleBinding](objs)
	require.Len(t, rbs, 2)
}

func Test_Generators_Permissions_EmptySANameTreatedAsDefault(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeOwnNamespace).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces("watch-namespace"))
	require.NoError(t, err)

	role := findObjectByKindAndName[*rbacv1.Role](objs, "csv-default")
	require.NotNil(t, role)
	require.Equal(t, "watch-namespace", role.Namespace)

	rb := findObjectByKindAndName[*rbacv1.RoleBinding](objs, "csv-default")
	require.NotNil(t, rb)
	require.Equal(t, []rbacv1.Subject{{Kind: "ServiceAccount", Name: "default", Namespace: "install-namespace"}}, rb.Subjects)
}

func Test_Generators_Permissions_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	roles := findObjectsByKind[*rbacv1.Role](objs)
	require.Empty(t, roles)
}

// ---------------------------------------------------------------------------
// ClusterPermissions Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_ClusterPermissions_PromotesInAllNamespacesMode(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-one",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-two",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces(""))
	require.NoError(t, err)

	crs := findObjectsByKind[*rbacv1.ClusterRole](objs)
	require.Len(t, crs, 2)

	cr1 := findObjectByKindAndName[*rbacv1.ClusterRole](objs, "csv-service-account-one")
	require.NotNil(t, cr1)
	// Should have original rule + namespace rule
	require.Len(t, cr1.Rules, 2)
	require.Equal(t, []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
		{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{corev1.GroupName}, Resources: []string{"namespaces"}},
	}, cr1.Rules)

	cr2 := findObjectByKindAndName[*rbacv1.ClusterRole](objs, "csv-service-account-two")
	require.NotNil(t, cr2)
	require.Len(t, cr2.Rules, 2)
	require.Equal(t, []rbacv1.PolicyRule{
		{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
		{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{corev1.GroupName}, Resources: []string{"namespaces"}},
	}, cr2.Rules)

	crbs := findObjectsByKind[*rbacv1.ClusterRoleBinding](objs)
	require.Len(t, crbs, 2)

	crb1 := findObjectByKindAndName[*rbacv1.ClusterRoleBinding](objs, "csv-service-account-one")
	require.NotNil(t, crb1)
	require.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "csv-service-account-one"}, crb1.RoleRef)
	require.Equal(t, []rbacv1.Subject{{Kind: "ServiceAccount", Name: "service-account-one", Namespace: "install-namespace"}}, crb1.Subjects)
}

func Test_Generators_ClusterPermissions_GeneratesClusterRolesAndBindings(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeOwnNamespace).
			WithClusterPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-one",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-two",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces("watch-namespace"))
	require.NoError(t, err)

	cr1 := findObjectByKindAndName[*rbacv1.ClusterRole](objs, "csv-service-account-one")
	require.NotNil(t, cr1)
	require.Equal(t, []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
	}, cr1.Rules)

	cr2 := findObjectByKindAndName[*rbacv1.ClusterRole](objs, "csv-service-account-two")
	require.NotNil(t, cr2)
	require.Equal(t, []rbacv1.PolicyRule{
		{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
	}, cr2.Rules)

	crb1 := findObjectByKindAndName[*rbacv1.ClusterRoleBinding](objs, "csv-service-account-one")
	require.NotNil(t, crb1)
	require.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "csv-service-account-one"}, crb1.RoleRef)
	require.Equal(t, []rbacv1.Subject{{Kind: "ServiceAccount", Name: "service-account-one", Namespace: "install-namespace"}}, crb1.Subjects)

	crb2 := findObjectByKindAndName[*rbacv1.ClusterRoleBinding](objs, "csv-service-account-two")
	require.NotNil(t, crb2)
	require.Equal(t, rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "csv-service-account-two"}, crb2.RoleRef)
	require.Equal(t, []rbacv1.Subject{{Kind: "ServiceAccount", Name: "service-account-two", Namespace: "install-namespace"}}, crb2.Subjects)
}

func Test_Generators_ClusterPermissions_EmptySAAsDefault(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeSingleNamespace, v1alpha1.InstallModeTypeOwnNamespace).
			WithClusterPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces("watch-namespace"))
	require.NoError(t, err)

	cr := findObjectByKindAndName[*rbacv1.ClusterRole](objs, "csv-default")
	require.NotNil(t, cr)

	crb := findObjectByKindAndName[*rbacv1.ClusterRoleBinding](objs, "csv-default")
	require.NotNil(t, crb)
	require.Equal(t, []rbacv1.Subject{{Kind: "ServiceAccount", Name: "default", Namespace: "install-namespace"}}, crb.Subjects)
}

func Test_Generators_ClusterPermissions_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	crs := findObjectsByKind[*rbacv1.ClusterRole](objs)
	require.Empty(t, crs)
}

// ---------------------------------------------------------------------------
// ServiceAccount Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_ServiceAccount_GeneratesUniqueSAs(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-1",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-2",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
					},
				},
			).
			WithClusterPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-2",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "service-account-3",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{"appsv1"}, Resources: []string{"deployments"}, Verbs: []string{"create"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	sas := findObjectsByKind[*corev1.ServiceAccount](objs)
	slices.SortFunc(sas, func(a, b *corev1.ServiceAccount) int {
		return cmp.Compare(a.Name, b.Name)
	})
	require.Len(t, sas, 3)
	require.Equal(t, "service-account-1", sas[0].Name)
	require.Equal(t, "install-namespace", sas[0].Namespace)
	require.Equal(t, "service-account-2", sas[1].Name)
	require.Equal(t, "install-namespace", sas[1].Namespace)
	require.Equal(t, "service-account-3", sas[2].Name)
	require.Equal(t, "install-namespace", sas[2].Namespace)
}

func Test_Generators_ServiceAccount_EmptySADefaultNotGenerated(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithName("csv").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
			).
			WithClusterPermissions(
				v1alpha1.StrategyDeploymentPermissions{
					ServiceAccountName: "",
					Rules: []rbacv1.PolicyRule{
						{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"get", "list", "watch"}},
					},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	sas := findObjectsByKind[*corev1.ServiceAccount](objs)
	require.Empty(t, sas)
}

func Test_Generators_ServiceAccount_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	sas := findObjectsByKind[*corev1.ServiceAccount](objs)
	require.Empty(t, sas)
}

// ---------------------------------------------------------------------------
// CRD Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_CRD_GeneratesCRDResources(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "crd-one"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "crd-two"}},
		},
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	crds := findObjectsByKind[*apiextensionsv1.CustomResourceDefinition](objs)
	require.Len(t, crds, 2)

	slices.SortFunc(crds, func(a, b *apiextensionsv1.CustomResourceDefinition) int {
		return cmp.Compare(a.Name, b.Name)
	})
	require.Equal(t, "crd-one", crds[0].Name)
	require.Equal(t, "crd-two", crds[1].Name)
}

func Test_Generators_CRD_WithConversionWebhook(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:                    v1alpha1.ConversionWebhook,
					GenerateName:            "crd-one-webhook",
					WebhookPath:             ptr.To("/some/path"),
					ContainerPort:           8443,
					AdmissionReviewVersions: []string{"v1", "v1beta1"},
					ConversionCRDs:          []string{"crd-one"},
					DeploymentName:          "some-deployment",
				},
				v1alpha1.WebhookDescription{
					Type:                    v1alpha1.ConversionWebhook,
					GenerateName:            "crd-two-webhook",
					ContainerPort:           8443,
					AdmissionReviewVersions: []string{"v1", "v1beta1"},
					ConversionCRDs:          []string{"crd-two"},
					DeploymentName:          "some-deployment",
				},
			).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{Name: "crd-one", Version: "v1", Kind: "CrdOne"},
				v1alpha1.CRDDescription{Name: "crd-two", Version: "v1", Kind: "CrdTwo"},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "some-deployment"},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "crd-one"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "crd-two"}},
		},
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	crds := findObjectsByKind[*apiextensionsv1.CustomResourceDefinition](objs)
	require.Len(t, crds, 2)

	slices.SortFunc(crds, func(a, b *apiextensionsv1.CustomResourceDefinition) int {
		return cmp.Compare(a.Name, b.Name)
	})

	crd1 := crds[0]
	require.Equal(t, "crd-one", crd1.Name)
	require.NotNil(t, crd1.Spec.Conversion)
	require.Equal(t, apiextensionsv1.WebhookConverter, crd1.Spec.Conversion.Strategy)
	require.NotNil(t, crd1.Spec.Conversion.Webhook)
	require.NotNil(t, crd1.Spec.Conversion.Webhook.ClientConfig)
	require.NotNil(t, crd1.Spec.Conversion.Webhook.ClientConfig.Service)
	require.Equal(t, "install-namespace", crd1.Spec.Conversion.Webhook.ClientConfig.Service.Namespace)
	require.Equal(t, "some-deployment-service", crd1.Spec.Conversion.Webhook.ClientConfig.Service.Name)
	require.Equal(t, ptr.To("/some/path"), crd1.Spec.Conversion.Webhook.ClientConfig.Service.Path)
	require.Equal(t, ptr.To(int32(8443)), crd1.Spec.Conversion.Webhook.ClientConfig.Service.Port)
	require.Equal(t, []string{"v1", "v1beta1"}, crd1.Spec.Conversion.Webhook.ConversionReviewVersions)

	crd2 := crds[1]
	require.Equal(t, "crd-two", crd2.Name)
	require.NotNil(t, crd2.Spec.Conversion)
	require.Equal(t, ptr.To("/"), crd2.Spec.Conversion.Webhook.ClientConfig.Service.Path)
}

func Test_Generators_CRD_PreserveUnknownFieldsFails(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:                    v1alpha1.ConversionWebhook,
					GenerateName:            "test-webhook",
					WebhookPath:             ptr.To("/some/path"),
					ContainerPort:           8443,
					AdmissionReviewVersions: []string{"v1", "v1beta1"},
					ConversionCRDs:          []string{"crd-one"},
					DeploymentName:          "some-deployment",
				},
			).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{Name: "crd-one", Version: "v1", Kind: "CrdOne"},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "some-deployment"},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "crd-one"},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					PreserveUnknownFields: true,
				},
			},
		},
	}

	_, err := renderer.Render(rv1, "install-namespace")
	require.Error(t, err)
	require.Contains(t, err.Error(), "must have .spec.preserveUnknownFields set to false to let API Server call webhook to do the conversion")
}

func Test_Generators_CRD_WithCertProvider(t *testing.T) {
	fakeProvider := render.FakeCertProvider{
		InjectCABundleFn: func(obj client.Object, _ render.CertificateProvisionerConfig) error {
			obj.SetAnnotations(map[string]string{
				"cert-provider": "annotation",
			})
			return nil
		},
		AdditionalObjectsFn: func(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return nil, nil
		},
		GetCertSecretInfoFn: func(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
			return render.CertSecretInfo{}
		},
	}

	renderer := buildRenderer(withCertProvider(fakeProvider))

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ConversionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
					ConversionCRDs: []string{"crd-one"},
				},
			).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{Name: "crd-one", Version: "v1", Kind: "CrdOne"},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "crd-one"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "crd-two"}},
		},
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	crds := findObjectsByKind[*apiextensionsv1.CustomResourceDefinition](objs)
	require.Len(t, crds, 2)

	// Find the CRD with conversion webhook (crd-one) - it should have cert annotations
	crd1 := findObjectByKindAndName[*apiextensionsv1.CustomResourceDefinition](objs, "crd-one")
	require.NotNil(t, crd1)
	require.Equal(t, map[string]string{"cert-provider": "annotation"}, crd1.GetAnnotations())
}

func Test_Generators_CRD_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	crds := findObjectsByKind[*apiextensionsv1.CustomResourceDefinition](objs)
	require.Empty(t, crds)
}

// ---------------------------------------------------------------------------
// Additional Resources Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_AdditionalResources_GeneratesResources(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		Others: []unstructured.Unstructured{
			*ToUnstructuredT(t,
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "bundled-service",
					},
				},
			),
			*ToUnstructuredT(t,
				&rbacv1.ClusterRole{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterRole",
						APIVersion: rbacv1.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "bundled-clusterrole",
					},
				},
			),
		},
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	// Find the additional resources (they come out as *unstructured.Unstructured)
	var unstructuredObjs []*unstructured.Unstructured
	for _, obj := range objs {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			unstructuredObjs = append(unstructuredObjs, u)
		}
	}
	require.Len(t, unstructuredObjs, 2)

	// Service should get namespace set (namespaced resource)
	svc := findObjectByKindAndName[*unstructured.Unstructured](objs, "bundled-service")
	require.NotNil(t, svc)
	require.Equal(t, "install-namespace", svc.GetNamespace())

	// ClusterRole should not get namespace set (cluster-scoped resource)
	cr := findObjectByKindAndName[*unstructured.Unstructured](objs, "bundled-clusterrole")
	require.NotNil(t, cr)
	// ClusterRoles are not namespaced, so namespace should be empty
	require.Empty(t, cr.GetNamespace())
}

func Test_Generators_AdditionalResources_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	var unstructuredObjs []*unstructured.Unstructured
	for _, obj := range objs {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			unstructuredObjs = append(unstructuredObjs, u)
		}
	}
	require.Empty(t, unstructuredObjs)
}

// ---------------------------------------------------------------------------
// Validating Webhook Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_ValidatingWebhook_GeneratesWebhookConfig(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ValidatingAdmissionWebhook,
					GenerateName:   "my-webhook",
					DeploymentName: "my-deployment",
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.OperationAll},
							Rule: admissionregistrationv1.Rule{
								APIGroups: []string{""}, APIVersions: []string{""}, Resources: []string{"namespaces"},
							},
						},
					},
					FailurePolicy: ptr.To(admissionregistrationv1.Fail),
					ObjectSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
					SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
					TimeoutSeconds:          ptr.To(int32(1)),
					AdmissionReviewVersions: []string{"v1beta1", "v1beta2"},
					WebhookPath:             ptr.To("/webhook-path"),
					ContainerPort:           443,
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces(""))
	require.NoError(t, err)

	vwcs := findObjectsByKind[*admissionregistrationv1.ValidatingWebhookConfiguration](objs)
	require.Len(t, vwcs, 1)

	vwc := vwcs[0]
	require.Equal(t, "my-webhook", vwc.Name)
	require.Equal(t, "install-namespace", vwc.Namespace)
	require.Len(t, vwc.Webhooks, 1)
	wh := vwc.Webhooks[0]
	require.Equal(t, "my-webhook", wh.Name)
	require.Equal(t, ptr.To(admissionregistrationv1.Fail), wh.FailurePolicy)
	require.Equal(t, &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}}, wh.ObjectSelector)
	require.Equal(t, ptr.To(admissionregistrationv1.SideEffectClassNone), wh.SideEffects)
	require.Equal(t, ptr.To(int32(1)), wh.TimeoutSeconds)
	require.Equal(t, []string{"v1beta1", "v1beta2"}, wh.AdmissionReviewVersions)
	require.NotNil(t, wh.ClientConfig.Service)
	require.Equal(t, "install-namespace", wh.ClientConfig.Service.Namespace)
	require.Equal(t, "my-deployment-service", wh.ClientConfig.Service.Name)
	require.Equal(t, ptr.To("/webhook-path"), wh.ClientConfig.Service.Path)
	require.Equal(t, ptr.To(int32(443)), wh.ClientConfig.Service.Port)
	// No NamespaceSelector in AllNamespaces mode
	require.Nil(t, wh.NamespaceSelector)
}

func Test_Generators_ValidatingWebhook_TrimsTrailingDash(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ValidatingAdmissionWebhook,
					GenerateName:   "my-webhook",
					DeploymentName: "my-deployment",
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.OperationAll},
							Rule: admissionregistrationv1.Rule{
								APIGroups: []string{""}, APIVersions: []string{""}, Resources: []string{"namespaces"},
							},
						},
					},
					FailurePolicy: ptr.To(admissionregistrationv1.Fail),
					ObjectSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
					SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
					TimeoutSeconds:          ptr.To(int32(1)),
					AdmissionReviewVersions: []string{"v1beta1", "v1beta2"},
					WebhookPath:             ptr.To("/webhook-path"),
					ContainerPort:           443,
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	vwc := findObjectByKindAndName[*admissionregistrationv1.ValidatingWebhookConfiguration](objs, "my-webhook")
	require.NotNil(t, vwc)
	require.Equal(t, "my-webhook", vwc.Webhooks[0].Name)

	// Namespace selector should be set for non-AllNamespaces mode
	require.NotNil(t, vwc.Webhooks[0].NamespaceSelector)
	require.Equal(t, &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "kubernetes.io/metadata.name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"watch-namespace-one", "watch-namespace-two"},
			},
		},
	}, vwc.Webhooks[0].NamespaceSelector)
}

func Test_Generators_ValidatingWebhook_WithCertProvider(t *testing.T) {
	fakeProvider := render.FakeCertProvider{
		InjectCABundleFn: func(obj client.Object, _ render.CertificateProvisionerConfig) error {
			obj.SetAnnotations(map[string]string{"cert-provider": "annotation"})
			return nil
		},
		AdditionalObjectsFn: func(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return nil, nil
		},
		GetCertSecretInfoFn: func(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
			return render.CertSecretInfo{}
		},
	}
	renderer := buildRenderer(withCertProvider(fakeProvider))

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ValidatingAdmissionWebhook,
					GenerateName:   "my-webhook",
					DeploymentName: "my-deployment",
					ContainerPort:  443,
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	vwc := findObjectByKindAndName[*admissionregistrationv1.ValidatingWebhookConfiguration](objs, "my-webhook")
	require.NotNil(t, vwc)
	require.Equal(t, map[string]string{"cert-provider": "annotation"}, vwc.GetAnnotations())
}

func Test_Generators_ValidatingWebhook_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	vwcs := findObjectsByKind[*admissionregistrationv1.ValidatingWebhookConfiguration](objs)
	require.Empty(t, vwcs)
}

// ---------------------------------------------------------------------------
// Mutating Webhook Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_MutatingWebhook_GeneratesWebhookConfig(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "my-webhook",
					DeploymentName: "my-deployment",
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.OperationAll},
							Rule: admissionregistrationv1.Rule{
								APIGroups: []string{""}, APIVersions: []string{""}, Resources: []string{"namespaces"},
							},
						},
					},
					FailurePolicy: ptr.To(admissionregistrationv1.Fail),
					ObjectSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
					SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
					TimeoutSeconds:          ptr.To(int32(1)),
					AdmissionReviewVersions: []string{"v1beta1", "v1beta2"},
					WebhookPath:             ptr.To("/webhook-path"),
					ContainerPort:           443,
					ReinvocationPolicy:      ptr.To(admissionregistrationv1.IfNeededReinvocationPolicy),
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithTargetNamespaces(""))
	require.NoError(t, err)

	mwcs := findObjectsByKind[*admissionregistrationv1.MutatingWebhookConfiguration](objs)
	require.Len(t, mwcs, 1)

	mwc := mwcs[0]
	require.Equal(t, "my-webhook", mwc.Name)
	require.Equal(t, "install-namespace", mwc.Namespace)
	require.Len(t, mwc.Webhooks, 1)
	wh := mwc.Webhooks[0]
	require.Equal(t, "my-webhook", wh.Name)
	require.Equal(t, ptr.To(admissionregistrationv1.Fail), wh.FailurePolicy)
	require.Equal(t, &metav1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}}, wh.ObjectSelector)
	require.Equal(t, ptr.To(admissionregistrationv1.SideEffectClassNone), wh.SideEffects)
	require.Equal(t, ptr.To(int32(1)), wh.TimeoutSeconds)
	require.Equal(t, []string{"v1beta1", "v1beta2"}, wh.AdmissionReviewVersions)
	require.Equal(t, ptr.To(admissionregistrationv1.IfNeededReinvocationPolicy), wh.ReinvocationPolicy)
	require.NotNil(t, wh.ClientConfig.Service)
	require.Equal(t, "install-namespace", wh.ClientConfig.Service.Namespace)
	require.Equal(t, "my-deployment-service", wh.ClientConfig.Service.Name)
	require.Equal(t, ptr.To("/webhook-path"), wh.ClientConfig.Service.Path)
	require.Equal(t, ptr.To(int32(443)), wh.ClientConfig.Service.Port)
	// No NamespaceSelector in AllNamespaces mode
	require.Nil(t, wh.NamespaceSelector)
}

func Test_Generators_MutatingWebhook_TrimsTrailingDash(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "my-webhook",
					DeploymentName: "my-deployment",
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.OperationAll},
							Rule: admissionregistrationv1.Rule{
								APIGroups: []string{""}, APIVersions: []string{""}, Resources: []string{"namespaces"},
							},
						},
					},
					FailurePolicy: ptr.To(admissionregistrationv1.Fail),
					ObjectSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
					SideEffects:             ptr.To(admissionregistrationv1.SideEffectClassNone),
					TimeoutSeconds:          ptr.To(int32(1)),
					AdmissionReviewVersions: []string{"v1beta1", "v1beta2"},
					WebhookPath:             ptr.To("/webhook-path"),
					ContainerPort:           443,
					ReinvocationPolicy:      ptr.To(admissionregistrationv1.IfNeededReinvocationPolicy),
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	mwc := findObjectByKindAndName[*admissionregistrationv1.MutatingWebhookConfiguration](objs, "my-webhook")
	require.NotNil(t, mwc)
	require.Equal(t, "my-webhook", mwc.Webhooks[0].Name)

	// Namespace selector should be set
	require.NotNil(t, mwc.Webhooks[0].NamespaceSelector)
	require.Equal(t, &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "kubernetes.io/metadata.name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"watch-namespace-one", "watch-namespace-two"},
			},
		},
	}, mwc.Webhooks[0].NamespaceSelector)
}

func Test_Generators_MutatingWebhook_WithCertProvider(t *testing.T) {
	fakeProvider := render.FakeCertProvider{
		InjectCABundleFn: func(obj client.Object, _ render.CertificateProvisionerConfig) error {
			obj.SetAnnotations(map[string]string{"cert-provider": "annotation"})
			return nil
		},
		AdditionalObjectsFn: func(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return nil, nil
		},
		GetCertSecretInfoFn: func(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
			return render.CertSecretInfo{}
		},
	}
	renderer := buildRenderer(withCertProvider(fakeProvider))

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "my-webhook",
					DeploymentName: "my-deployment",
					ContainerPort:  443,
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	mwc := findObjectByKindAndName[*admissionregistrationv1.MutatingWebhookConfiguration](objs, "my-webhook")
	require.NotNil(t, mwc)
	require.Equal(t, map[string]string{"cert-provider": "annotation"}, mwc.GetAnnotations())
}

func Test_Generators_MutatingWebhook_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)
	mwcs := findObjectsByKind[*admissionregistrationv1.MutatingWebhookConfiguration](objs)
	require.Empty(t, mwcs)
}

// ---------------------------------------------------------------------------
// Deployment Service Resource Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_DeploymentService_DefaultPort(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Equal(t, "install-namespace", svc.Namespace)
	require.Len(t, svc.Spec.Ports, 1)
	require.Equal(t, int32(443), svc.Spec.Ports[0].Port)
	require.Equal(t, intstr.FromInt32(443), svc.Spec.Ports[0].TargetPort)
	require.Equal(t, "443", svc.Spec.Ports[0].Name)
}

func Test_Generators_DeploymentService_CustomContainerPort(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ValidatingAdmissionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
					ContainerPort:  int32(8443),
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Len(t, svc.Spec.Ports, 1)
	require.Equal(t, int32(8443), svc.Spec.Ports[0].Port)
	require.Equal(t, intstr.FromInt32(8443), svc.Spec.Ports[0].TargetPort)
}

func Test_Generators_DeploymentService_CustomTargetPort(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ConversionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
					TargetPort:     &intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces(""),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Len(t, svc.Spec.Ports, 1)
	require.Equal(t, int32(443), svc.Spec.Ports[0].Port)
	require.Equal(t, intstr.IntOrString{Type: intstr.Int, IntVal: 8080}, svc.Spec.Ports[0].TargetPort)
}

func Test_Generators_DeploymentService_ContainerAndTargetPort(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ConversionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
					ContainerPort:  int32(9090),
					TargetPort:     &intstr.IntOrString{Type: intstr.Int, IntVal: 9099},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces(""),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Len(t, svc.Spec.Ports, 1)
	require.Equal(t, int32(9090), svc.Spec.Ports[0].Port)
	require.Equal(t, intstr.IntOrString{Type: intstr.Int, IntVal: 9099}, svc.Spec.Ports[0].TargetPort)
}

func Test_Generators_DeploymentService_UsesDeploymentLabelSelector(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{
					Name: "my-deployment",
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"foo": "bar"},
						},
					},
				},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ConversionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
					ContainerPort:  int32(9090),
					TargetPort:     &intstr.IntOrString{Type: intstr.Int, IntVal: 9099},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces(""),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Equal(t, map[string]string{"foo": "bar"}, svc.Spec.Selector)
}

func Test_Generators_DeploymentService_AggregatesMultipleWebhooks(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{
					Name: "my-deployment",
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"foo": "bar"},
						},
					},
				},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "mutating-webhook",
					DeploymentName: "my-deployment",
				},
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ValidatingAdmissionWebhook,
					GenerateName:   "validating-webhook",
					DeploymentName: "my-deployment",
					ContainerPort:  int32(8443),
				},
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ConversionWebhook,
					GenerateName:   "conversion-webhook-1",
					DeploymentName: "my-deployment",
					TargetPort:     &intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.ConversionWebhook,
					GenerateName:   "conversion-webhook-2",
					DeploymentName: "my-deployment",
					ContainerPort:  int32(9090),
					TargetPort:     &intstr.IntOrString{Type: intstr.Int, IntVal: 9099},
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces(""),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Len(t, svc.Spec.Ports, 4)
	require.Equal(t, map[string]string{"foo": "bar"}, svc.Spec.Selector)

	// Verify the ports are sorted by Port then TargetPort
	require.Equal(t, []corev1.ServicePort{
		{Name: "443", Port: int32(443), TargetPort: intstr.FromInt32(443)},
		{Name: "443", Port: int32(443), TargetPort: intstr.FromInt32(8080)},
		{Name: "8443", Port: int32(8443), TargetPort: intstr.FromInt32(8443)},
		{Name: "9090", Port: int32(9090), TargetPort: intstr.FromInt32(9099)},
	}, svc.Spec.Ports)
}

func Test_Generators_DeploymentService_WithCertProvider(t *testing.T) {
	fakeProvider := render.FakeCertProvider{
		InjectCABundleFn: func(obj client.Object, _ render.CertificateProvisionerConfig) error {
			obj.SetAnnotations(map[string]string{"cert-provider": "annotation"})
			return nil
		},
		AdditionalObjectsFn: func(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return nil, nil
		},
		GetCertSecretInfoFn: func(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
			return render.CertSecretInfo{}
		},
	}

	renderer := buildRenderer(withCertProvider(fakeProvider))

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeMultiNamespace).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
			).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
				},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace",
		render.WithTargetNamespaces("watch-namespace-one", "watch-namespace-two"),
	)
	require.NoError(t, err)

	svc := findObjectByKindAndName[*corev1.Service](objs, "my-deployment-service")
	require.NotNil(t, svc)
	require.Equal(t, map[string]string{"cert-provider": "annotation"}, svc.GetAnnotations())
}

func Test_Generators_DeploymentService_EmptyBundle(t *testing.T) {
	renderer := buildRenderer()
	objs, err := renderer.Render(bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
	}, "install-namespace")
	require.NoError(t, err)

	// Without webhooks, no service should be generated
	svcs := findObjectsByKind[*corev1.Service](objs)
	require.Empty(t, svcs)
}

// ---------------------------------------------------------------------------
// Cert Provider Resource Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_CertProviderResource_GeneratesCertResources(t *testing.T) {
	fakeProvider := render.FakeCertProvider{
		AdditionalObjectsFn: func(cfg render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
			return []unstructured.Unstructured{*ToUnstructuredT(t, &corev1.Secret{
				TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()},
				ObjectMeta: metav1.ObjectMeta{Name: cfg.CertName},
			})}, nil
		},
		InjectCABundleFn: func(_ client.Object, _ render.CertificateProvisionerConfig) error {
			return nil
		},
		GetCertSecretInfoFn: func(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
			return render.CertSecretInfo{}
		},
	}

	renderer := buildRenderer(withCertProvider(fakeProvider))

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithWebhookDefinitions(
				v1alpha1.WebhookDescription{
					Type:           v1alpha1.MutatingAdmissionWebhook,
					GenerateName:   "test-webhook",
					DeploymentName: "my-deployment",
				},
			).
			WithStrategyDeploymentSpecs(
				v1alpha1.StrategyDeploymentSpec{Name: "my-deployment"},
				v1alpha1.StrategyDeploymentSpec{Name: "my-other-deployment"},
			).Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	// Find unstructured Secrets added by cert provider
	var certSecrets []*unstructured.Unstructured
	for _, obj := range objs {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			if u.GetKind() == "Secret" {
				certSecrets = append(certSecrets, u)
			}
		}
	}
	require.Len(t, certSecrets, 1)
	require.Equal(t, "my-deployment-service-cert", certSecrets[0].GetName())
}

// ---------------------------------------------------------------------------
// Provided APIs ClusterRoles Generator Tests
// ---------------------------------------------------------------------------

func Test_Generators_ProvidedAPIs_SingleCRDGenerates4ClusterRoles(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{
					Name:    "etcdclusters.etcd.database.coreos.com",
					Version: "v1beta2",
					Kind:    "EtcdCluster",
				},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "etcdclusters.etcd.database.coreos.com"}},
		},
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithProvidedAPIsClusterRoles())
	require.NoError(t, err)

	// Find generated provided API cluster roles (not the ones from clusterPermissions)
	var apiClusterRoles []*rbacv1.ClusterRole
	for _, obj := range objs {
		if cr, ok := obj.(*rbacv1.ClusterRole); ok {
			// Filter to only the providedAPIs generated ones by checking the name prefix
			if len(cr.Name) > 0 && (cr.Name == "etcdclusters.etcd.database.coreos.com-v1beta2-admin" ||
				cr.Name == "etcdclusters.etcd.database.coreos.com-v1beta2-edit" ||
				cr.Name == "etcdclusters.etcd.database.coreos.com-v1beta2-view" ||
				cr.Name == "etcdclusters.etcd.database.coreos.com-v1beta2-crd-view") {
				apiClusterRoles = append(apiClusterRoles, cr)
			}
		}
	}
	require.Len(t, apiClusterRoles, 4)

	rolesByName := map[string]*rbacv1.ClusterRole{}
	for _, cr := range apiClusterRoles {
		rolesByName[cr.Name] = cr
	}

	admin := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-admin"]
	require.NotNil(t, admin)
	assert.Equal(t, "true", admin.Labels["rbac.authorization.k8s.io/aggregate-to-admin"])
	assert.Equal(t, []rbacv1.PolicyRule{{
		Verbs:     []string{"*"},
		APIGroups: []string{"etcd.database.coreos.com"},
		Resources: []string{"etcdclusters"},
	}}, admin.Rules)

	edit := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-edit"]
	require.NotNil(t, edit)
	assert.Equal(t, "true", edit.Labels["rbac.authorization.k8s.io/aggregate-to-edit"])
	assert.Equal(t, []rbacv1.PolicyRule{{
		Verbs:     []string{"create", "update", "patch", "delete"},
		APIGroups: []string{"etcd.database.coreos.com"},
		Resources: []string{"etcdclusters"},
	}}, edit.Rules)

	view := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-view"]
	require.NotNil(t, view)
	assert.Equal(t, "true", view.Labels["rbac.authorization.k8s.io/aggregate-to-view"])
	assert.Equal(t, []rbacv1.PolicyRule{{
		Verbs:     []string{"get", "list", "watch"},
		APIGroups: []string{"etcd.database.coreos.com"},
		Resources: []string{"etcdclusters"},
	}}, view.Rules)

	crdView := rolesByName["etcdclusters.etcd.database.coreos.com-v1beta2-crd-view"]
	require.NotNil(t, crdView)
	assert.Equal(t, "true", crdView.Labels["rbac.authorization.k8s.io/aggregate-to-view"])
	assert.Equal(t, []rbacv1.PolicyRule{{
		Verbs:         []string{"get"},
		APIGroups:     []string{"apiextensions.k8s.io"},
		Resources:     []string{"customresourcedefinitions"},
		ResourceNames: []string{"etcdclusters.etcd.database.coreos.com"},
	}}, crdView.Rules)
}

func Test_Generators_ProvidedAPIs_MultipleCRDsGenerate4Each(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{Name: "foos.example.com", Version: "v1", Kind: "Foo"},
				v1alpha1.CRDDescription{Name: "bars.example.com", Version: "v1", Kind: "Bar"},
				v1alpha1.CRDDescription{Name: "bazzes.other.io", Version: "v2alpha1", Kind: "Baz"},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "foos.example.com"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "bars.example.com"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "bazzes.other.io"}},
		},
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithProvidedAPIsClusterRoles())
	require.NoError(t, err)

	// Count cluster roles with aggregate labels (to filter out any from clusterPermissions)
	var apiClusterRoles []*rbacv1.ClusterRole
	for _, obj := range objs {
		if cr, ok := obj.(*rbacv1.ClusterRole); ok {
			for key := range cr.Labels {
				if key == "rbac.authorization.k8s.io/aggregate-to-admin" ||
					key == "rbac.authorization.k8s.io/aggregate-to-edit" ||
					key == "rbac.authorization.k8s.io/aggregate-to-view" {
					apiClusterRoles = append(apiClusterRoles, cr)
					break
				}
			}
		}
	}
	assert.Len(t, apiClusterRoles, 12)
}

func Test_Generators_ProvidedAPIs_NoOwnedCRDsGeneratesEmpty(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithOwnedCRDs().Build(),
	}

	objs, err := renderer.Render(rv1, "install-namespace", render.WithProvidedAPIsClusterRoles())
	require.NoError(t, err)

	// Should have no cluster roles at all
	crs := findObjectsByKind[*rbacv1.ClusterRole](objs)
	assert.Empty(t, crs)
}

func Test_Generators_ProvidedAPIs_InvalidCRDNameReturnsError(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{Name: "invalidname", Version: "v1", Kind: "Invalid"},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "invalidname"}},
		},
	}

	_, err := renderer.Render(rv1, "install-namespace", render.WithProvidedAPIsClusterRoles())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CRD name")
}

func Test_Generators_ProvidedAPIs_NotGeneratedByDefault(t *testing.T) {
	renderer := buildRenderer()

	rv1 := bundle.RegistryV1{
		PackageName: "test-package",
		CSV: clusterserviceversion.Builder().
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).
			WithOwnedCRDs(
				v1alpha1.CRDDescription{Name: "foos.example.com", Version: "v1", Kind: "Foo"},
			).Build(),
		CRDs: []apiextensionsv1.CustomResourceDefinition{
			{ObjectMeta: metav1.ObjectMeta{Name: "foos.example.com"}},
		},
	}

	// Without WithProvidedAPIsClusterRoles(), no provided API cluster roles should be generated
	objs, err := renderer.Render(rv1, "install-namespace")
	require.NoError(t, err)

	crs := findObjectsByKind[*rbacv1.ClusterRole](objs)
	assert.Empty(t, crs)
}

// ---------------------------------------------------------------------------
// Deployment Config Tests
// ---------------------------------------------------------------------------

func Test_Generators_DeploymentConfig(t *testing.T) {
	for _, tc := range []struct {
		name             string
		csv              v1alpha1.ClusterServiceVersion
		deploymentConfig *render.DeploymentConfig
		verify           func(*testing.T, []client.Object)
	}{
		{
			name: "applies env vars from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "manager",
											Env:  []corev1.EnvVar{{Name: "EXISTING_VAR", Value: "existing_value"}},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Env: []corev1.EnvVar{
					{Name: "NEW_VAR", Value: "new_value"},
					{Name: "EXISTING_VAR", Value: "overridden_value"},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				envVars := dep.Spec.Template.Spec.Containers[0].Env
				require.Len(t, envVars, 2)

				var existingVar *corev1.EnvVar
				for i := range envVars {
					if envVars[i].Name == "EXISTING_VAR" {
						existingVar = &envVars[i]
						break
					}
				}
				require.NotNil(t, existingVar)
				require.Equal(t, "overridden_value", existingVar.Value)

				var newVar *corev1.EnvVar
				for i := range envVars {
					if envVars[i].Name == "NEW_VAR" {
						newVar = &envVars[i]
						break
					}
				}
				require.NotNil(t, newVar)
				require.Equal(t, "new_value", newVar.Value)
			},
		},
		{
			name: "applies resources from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Resources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				resources := dep.Spec.Template.Spec.Containers[0].Resources
				require.Equal(t, resource.MustParse("100m"), *resources.Requests.Cpu())
				require.Equal(t, resource.MustParse("128Mi"), *resources.Requests.Memory())
				require.Equal(t, resource.MustParse("200m"), *resources.Limits.Cpu())
				require.Equal(t, resource.MustParse("256Mi"), *resources.Limits.Memory())
			},
		},
		{
			name: "applies tolerations from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Tolerations: []corev1.Toleration{
					{
						Key:      "node.kubernetes.io/disk-type",
						Operator: corev1.TolerationOpEqual,
						Value:    "ssd",
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				tolerations := dep.Spec.Template.Spec.Tolerations
				require.Len(t, tolerations, 1)
				require.Equal(t, "node.kubernetes.io/disk-type", tolerations[0].Key)
				require.Equal(t, corev1.TolerationOpEqual, tolerations[0].Operator)
				require.Equal(t, "ssd", tolerations[0].Value)
				require.Equal(t, corev1.TaintEffectNoSchedule, tolerations[0].Effect)
			},
		},
		{
			name: "applies node selector from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers:   []corev1.Container{{Name: "manager"}},
									NodeSelector: map[string]string{"existing-key": "existing-value"},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				NodeSelector: map[string]string{"disk-type": "ssd"},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.Equal(t, map[string]string{"disk-type": "ssd"}, dep.Spec.Template.Spec.NodeSelector)
			},
		},
		{
			name: "applies affinity from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{Key: "kubernetes.io/arch", Operator: corev1.NodeSelectorOpIn, Values: []string{"amd64", "arm64"}},
									},
								},
							},
						},
					},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
				require.Len(t, dep.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 1)
			},
		},
		{
			name: "empty affinity erases bundle affinity",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Affinity: &corev1.Affinity{
										NodeAffinity: &corev1.NodeAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
												NodeSelectorTerms: []corev1.NodeSelectorTerm{
													{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
												},
											},
										},
										PodAffinity: &corev1.PodAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
												{LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"cache"}}}}, TopologyKey: "kubernetes.io/hostname"},
											},
										},
										PodAntiAffinity: &corev1.PodAntiAffinity{
											PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
												{Weight: 100, PodAffinityTerm: corev1.PodAffinityTerm{LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"database"}}}}, TopologyKey: "kubernetes.io/hostname"}},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.Nil(t, dep.Spec.Template.Spec.Affinity)
			},
		},
		{
			name: "empty nodeAffinity erases only nodeAffinity",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Affinity: &corev1.Affinity{
										NodeAffinity: &corev1.NodeAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
												NodeSelectorTerms: []corev1.NodeSelectorTerm{
													{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
												},
											},
										},
										PodAffinity: &corev1.PodAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
												{LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"cache"}}}}, TopologyKey: "kubernetes.io/hostname"},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity)
				require.Nil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.PodAffinity)
			},
		},
		{
			name: "empty nodeAffinity with empty nodeSelectorTerms erases only nodeAffinity",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Affinity: &corev1.Affinity{
										NodeAffinity: &corev1.NodeAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
												NodeSelectorTerms: []corev1.NodeSelectorTerm{
													{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
												},
											},
										},
										PodAffinity: &corev1.PodAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
												{LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"cache"}}}}, TopologyKey: "kubernetes.io/hostname"},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{},
						},
					},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity)
				require.Nil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity,
					"nodeAffinity should be erased when RequiredDuringSchedulingIgnoredDuringExecution has empty NodeSelectorTerms")
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.PodAffinity)
			},
		},
		{
			name: "nil affinity preserves bundle affinity",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Affinity: &corev1.Affinity{
										NodeAffinity: &corev1.NodeAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
												NodeSelectorTerms: []corev1.NodeSelectorTerm{
													{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
												},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
			},
		},
		{
			name: "partial affinity override preserves unspecified fields",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Affinity: &corev1.Affinity{
										NodeAffinity: &corev1.NodeAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
												NodeSelectorTerms: []corev1.NodeSelectorTerm{
													{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
												},
											},
										},
										PodAffinity: &corev1.PodAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
												{LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"cache"}}}}, TopologyKey: "kubernetes.io/hostname"},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "kubernetes.io/arch", Operator: corev1.NodeSelectorOpIn, Values: []string{"arm64"}}}},
							},
						},
					},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.NodeAffinity)
				require.Equal(t, "kubernetes.io/arch",
					dep.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)
				require.NotNil(t, dep.Spec.Template.Spec.Affinity.PodAffinity,
					"podAffinity should be preserved when not specified in config")
			},
		},
		{
			name: "empty sub-fields erase to nil affinity",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Affinity: &corev1.Affinity{
										NodeAffinity: &corev1.NodeAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
												NodeSelectorTerms: []corev1.NodeSelectorTerm{
													{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "disktype", Operator: corev1.NodeSelectorOpIn, Values: []string{"ssd"}}}},
												},
											},
										},
										PodAffinity: &corev1.PodAffinity{
											RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
												{LabelSelector: &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"cache"}}}}, TopologyKey: "kubernetes.io/hostname"},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{},
					PodAffinity:  &corev1.PodAffinity{},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.Nil(t, dep.Spec.Template.Spec.Affinity,
					"affinity should be nil when all sub-fields are erased")
			},
		},
		{
			name: "empty nodeAffinity without bundle affinity stays nil",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.Nil(t, dep.Spec.Template.Spec.Affinity,
					"empty sub-field should not create an affinity object when bundle has none")
			},
		},
		{
			name: "empty affinity without bundle affinity stays nil",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Affinity: &corev1.Affinity{},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.Nil(t, dep.Spec.Template.Spec.Affinity,
					"empty affinity should not create an affinity object when bundle has none")
			},
		},
		{
			name: "applies annotations from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithAnnotations(map[string]string{"csv-annotation": "csv-value"}).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Annotations: map[string]string{"existing-pod-annotation": "existing-pod-value"},
								},
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Annotations: map[string]string{
					"config-annotation":       "config-value",
					"existing-pod-annotation": "should-not-override",
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)

				require.Contains(t, dep.Annotations, "config-annotation")
				require.Equal(t, "config-value", dep.Annotations["config-annotation"])

				require.Contains(t, dep.Spec.Template.Annotations, "csv-annotation")
				require.Equal(t, "csv-value", dep.Spec.Template.Annotations["csv-annotation"])
				require.Contains(t, dep.Spec.Template.Annotations, "existing-pod-annotation")
				require.Equal(t, "existing-pod-value", dep.Spec.Template.Annotations["existing-pod-annotation"])
				require.Contains(t, dep.Spec.Template.Annotations, "config-annotation")
				require.Equal(t, "config-value", dep.Spec.Template.Annotations["config-annotation"])
			},
		},
		{
			name: "applies volumes and volume mounts from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Volumes: []corev1.Volume{
					{
						Name: "config-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
							},
						},
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "config-volume", MountPath: "/etc/config"},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)

				require.Len(t, dep.Spec.Template.Spec.Volumes, 1)
				require.Equal(t, "config-volume", dep.Spec.Template.Spec.Volumes[0].Name)

				require.Len(t, dep.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
				require.Equal(t, "config-volume", dep.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
				require.Equal(t, "/etc/config", dep.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
			},
		},
		{
			name: "applies envFrom from deployment config",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				EnvFrom: []corev1.EnvFromSource{
					{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "env-config"}}},
					{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "env-secret"}}},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)

				envFrom := dep.Spec.Template.Spec.Containers[0].EnvFrom
				require.Len(t, envFrom, 2)
				require.NotNil(t, envFrom[0].ConfigMapRef)
				require.Equal(t, "env-config", envFrom[0].ConfigMapRef.Name)
				require.NotNil(t, envFrom[1].SecretRef)
				require.Equal(t, "env-secret", envFrom[1].SecretRef.Name)
			},
		},
		{
			name: "applies all config fields together",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Env:       []corev1.EnvVar{{Name: "ENV_VAR", Value: "value"}},
				Resources: &corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}},
				Tolerations: []corev1.Toleration{
					{Key: "key1", Operator: corev1.TolerationOpEqual, Value: "value1"},
				},
				NodeSelector: map[string]string{"disk": "ssd"},
				Annotations:  map[string]string{"annotation-key": "annotation-value"},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)

				require.Len(t, dep.Spec.Template.Spec.Containers[0].Env, 1)
				require.Equal(t, "ENV_VAR", dep.Spec.Template.Spec.Containers[0].Env[0].Name)
				require.NotNil(t, dep.Spec.Template.Spec.Containers[0].Resources.Requests)
				require.Len(t, dep.Spec.Template.Spec.Tolerations, 1)
				require.Equal(t, map[string]string{"disk": "ssd"}, dep.Spec.Template.Spec.NodeSelector)
				require.Contains(t, dep.Annotations, "annotation-key")
				require.Contains(t, dep.Spec.Template.Annotations, "annotation-key")
			},
		},
		{
			name: "applies config to multiple containers",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "container1"},
										{Name: "container2"},
										{Name: "container3"},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Env:       []corev1.EnvVar{{Name: "SHARED_VAR", Value: "shared_value"}},
				Resources: &corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)

				for i := range dep.Spec.Template.Spec.Containers {
					container := dep.Spec.Template.Spec.Containers[i]
					require.Len(t, container.Env, 1)
					require.Equal(t, "SHARED_VAR", container.Env[0].Name)
					require.Equal(t, "shared_value", container.Env[0].Value)
					require.NotNil(t, container.Resources.Requests)
					require.Equal(t, resource.MustParse("100m"), *container.Resources.Requests.Cpu())
				}
			},
		},
		{
			name: "nil deployment config does nothing",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "manager",
											Env:  []corev1.EnvVar{{Name: "EXISTING_VAR", Value: "existing_value"}},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: nil,
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				require.Len(t, dep.Spec.Template.Spec.Containers[0].Env, 1)
				require.Equal(t, "EXISTING_VAR", dep.Spec.Template.Spec.Containers[0].Env[0].Name)
				require.Equal(t, "existing_value", dep.Spec.Template.Spec.Containers[0].Env[0].Value)
			},
		},
		{
			name: "merges volumes from deployment config - overrides matching names",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{Name: "manager"}},
									Volumes: []corev1.Volume{
										{Name: "bundle-emptydir-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
										{Name: "existing-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				Volumes: []corev1.Volume{
					{Name: "bundle-emptydir-vol", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "test-cm-vol"}}}},
					{Name: "config-secret-vol", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "test-secret-vol"}}},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				volumes := dep.Spec.Template.Spec.Volumes
				require.Len(t, volumes, 3)

				var bundleVol *corev1.Volume
				for i := range volumes {
					if volumes[i].Name == "bundle-emptydir-vol" {
						bundleVol = &volumes[i]
						break
					}
				}
				require.NotNil(t, bundleVol, "bundle-emptydir-vol should exist")
				require.NotNil(t, bundleVol.ConfigMap, "bundle-emptydir-vol should be ConfigMap")
				require.Equal(t, "test-cm-vol", bundleVol.ConfigMap.Name)
				require.Nil(t, bundleVol.EmptyDir, "bundle-emptydir-vol should not be EmptyDir")

				var existingVol *corev1.Volume
				for i := range volumes {
					if volumes[i].Name == "existing-vol" {
						existingVol = &volumes[i]
						break
					}
				}
				require.NotNil(t, existingVol, "existing-vol should exist")
				require.NotNil(t, existingVol.EmptyDir, "existing-vol should still be EmptyDir")

				var secretVol *corev1.Volume
				for i := range volumes {
					if volumes[i].Name == "config-secret-vol" {
						secretVol = &volumes[i]
						break
					}
				}
				require.NotNil(t, secretVol, "config-secret-vol should exist")
				require.NotNil(t, secretVol.Secret, "config-secret-vol should be Secret")
				require.Equal(t, "test-secret-vol", secretVol.Secret.SecretName)
			},
		},
		{
			name: "merges volumeMounts from deployment config - overrides matching names",
			csv: clusterserviceversion.Builder().
				WithInstallModeSupportFor(v1alpha1.InstallModeTypeOwnNamespace).
				WithStrategyDeploymentSpecs(
					v1alpha1.StrategyDeploymentSpec{
						Name: "test-deployment",
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "manager",
											VolumeMounts: []corev1.VolumeMount{
												{Name: "bundle-vol", MountPath: "/old/path"},
												{Name: "existing-vol", MountPath: "/existing/path"},
											},
										},
									},
								},
							},
						},
					},
				).Build(),
			deploymentConfig: &render.DeploymentConfig{
				VolumeMounts: []corev1.VolumeMount{
					{Name: "bundle-vol", MountPath: "/new/path"},
					{Name: "config-vol", MountPath: "/config/path"},
				},
			},
			verify: func(t *testing.T, objs []client.Object) {
				dep := findObjectByKindAndName[*appsv1.Deployment](objs, "test-deployment")
				require.NotNil(t, dep)
				volumeMounts := dep.Spec.Template.Spec.Containers[0].VolumeMounts
				require.Len(t, volumeMounts, 3)

				var bundleMount *corev1.VolumeMount
				for i := range volumeMounts {
					if volumeMounts[i].Name == "bundle-vol" {
						bundleMount = &volumeMounts[i]
						break
					}
				}
				require.NotNil(t, bundleMount, "bundle-vol should exist")
				require.Equal(t, "/new/path", bundleMount.MountPath, "bundle-vol mount path should be overridden")

				var existingMount *corev1.VolumeMount
				for i := range volumeMounts {
					if volumeMounts[i].Name == "existing-vol" {
						existingMount = &volumeMounts[i]
						break
					}
				}
				require.NotNil(t, existingMount, "existing-vol should exist")
				require.Equal(t, "/existing/path", existingMount.MountPath)

				var configMount *corev1.VolumeMount
				for i := range volumeMounts {
					if volumeMounts[i].Name == "config-vol" {
						configMount = &volumeMounts[i]
						break
					}
				}
				require.NotNil(t, configMount, "config-vol should exist")
				require.Equal(t, "/config/path", configMount.MountPath)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			renderer := buildRenderer(withDeploymentConfig(tc.deploymentConfig))
			objs, err := renderer.Render(
				bundle.RegistryV1{
					PackageName: "test-package", CSV: tc.csv},
				"test-ns",
				render.WithTargetNamespaces("test-ns"),
			)
			require.NoError(t, err)
			tc.verify(t, objs)
		})
	}
}

// ---------------------------------------------------------------------------
// Renderer Orchestration Tests
// ---------------------------------------------------------------------------

func Test_Renderer_NoConfig(t *testing.T) {
	renderer := render.NewRendererBuilder().Build()
	objs, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "")
	require.NoError(t, err)
	require.Empty(t, objs)
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
			renderer := render.NewRendererBuilder().Build()
			_, err := renderer.Render(bundle.RegistryV1{
				PackageName: "test-package",
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
			renderer := render.NewRendererBuilder().Build()
			_, err := renderer.Render(
				bundle.RegistryV1{
					PackageName: "test-package", CSV: tc.csv},
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

	renderer := render.NewRendererBuilder().WithDeploymentConfig(expectedConfig).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		},
		"test-namespace",
	)
	require.NoError(t, err)
}

func Test_Renderer_DeploymentConfig_NilWhenNotProvided(t *testing.T) {
	renderer := render.NewRendererBuilder().Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
}

func Test_Renderer_WithCertificateProvider(t *testing.T) {
	renderer := render.NewRendererBuilder().WithCertificateProvider(rendererTestCertProvider{}).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
}

func Test_Renderer_WithUniqueNameGenerator(t *testing.T) {
	customGen := func(base string, obj interface{}) string { return "custom-name" }
	renderer := render.NewRendererBuilder().WithUniqueNameGenerator(customGen).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
}

func Test_Renderer_DefaultOptionsPassedToGenerators(t *testing.T) {
	renderer := render.NewRendererBuilder().Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "install-ns")
	require.NoError(t, err)
}

func Test_Renderer_DeploymentConfig_NilWhenExplicitlyNil(t *testing.T) {
	renderer := render.NewRendererBuilder().WithDeploymentConfig(nil).Build()

	_, err := renderer.Render(
		bundle.RegistryV1{
			PackageName: "test-package",
			CSV:         clusterserviceversion.Builder().WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		}, "test-namespace")
	require.NoError(t, err)
}

type rendererTestCertProvider struct{}

func (f rendererTestCertProvider) InjectCABundle(_ client.Object, _ render.CertificateProvisionerConfig) error {
	return nil
}
func (f rendererTestCertProvider) AdditionalObjects(_ render.CertificateProvisionerConfig) ([]unstructured.Unstructured, error) {
	return nil, nil
}
func (f rendererTestCertProvider) GetCertSecretInfo(_ render.CertificateProvisionerConfig) render.CertSecretInfo {
	return render.CertSecretInfo{}
}

func Test_Renderer_Success(t *testing.T) {
	someBundle := bundle.RegistryV1{
		PackageName: "my-package",
		CSV: clusterserviceversion.Builder().
			WithName("test-bundle").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		Others: []unstructured.Unstructured{
			*ToUnstructuredT(t, &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-service",
				},
			}),
		},
	}

	renderer := render.NewRendererBuilder().Build()
	objs, err := renderer.Render(someBundle, "install-namespace")
	require.NoError(t, err)
	require.NotEmpty(t, objs)
	require.Len(t, objs, 1)
	require.Equal(t, "my-service", objs[0].GetName())
	require.Equal(t, "install-namespace", objs[0].GetNamespace())
}

func Test_Renderer_Failure_UnsupportedKind(t *testing.T) {
	someBundle := bundle.RegistryV1{
		PackageName: "my-package",
		CSV: clusterserviceversion.Builder().
			WithName("test-bundle").
			WithInstallModeSupportFor(v1alpha1.InstallModeTypeAllNamespaces).Build(),
		Others: []unstructured.Unstructured{
			*ToUnstructuredT(t, &corev1.Event{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Event",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testEvent",
				},
			}),
		},
	}

	renderer := render.NewRendererBuilder().Build()
	objs, err := renderer.Render(someBundle, "install-namespace")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported resource")
	require.Empty(t, objs)
}
