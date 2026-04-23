package render

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	registrybundle "github.com/operator-framework/operator-registry/pkg/lib/bundle"

	"github.com/perdasilva/regv1render/internal/bundle"
	"github.com/perdasilva/regv1render/internal/render/resourceutil"
)

const (
	labelKubernetesNamespaceMetadataName = "kubernetes.io/metadata.name"
)

type certVolumeConfig struct {
	Name        string
	Path        string
	TLSCertPath string
	TLSKeyPath  string
}

var certVolumeConfigs = []certVolumeConfig{
	{
		Name:        "webhook-cert",
		Path:        "/tmp/k8s-webhook-server/serving-certs",
		TLSCertPath: "tls.crt",
		TLSKeyPath:  "tls.key",
	}, {
		Name:        "apiservice-cert",
		Path:        "/apiserver.local.config/certificates",
		TLSCertPath: "apiserver.crt",
		TLSKeyPath:  "apiserver.key",
	},
}

func bundleCSVDeploymentGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}

	webhookDeployments := sets.Set[string]{}
	for _, wh := range rv1.CSV.Spec.WebhookDefinitions {
		webhookDeployments.Insert(wh.DeploymentName)
	}

	objs := make([]client.Object, 0, len(rv1.CSV.Spec.InstallStrategy.StrategySpec.DeploymentSpecs))
	for _, depSpec := range rv1.CSV.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		annotations := MergeMaps(rv1.CSV.Annotations, depSpec.Spec.Template.Annotations)
		annotations["olm.targetNamespaces"] = strings.Join(opts.TargetNamespaces, ",")
		depSpec.Spec.Template.Annotations = annotations
		depSpec.Spec.RevisionHistoryLimit = ptr.To(int32(1))

		deploymentResource := resourceutil.CreateDeploymentResource(
			depSpec.Name,
			opts.InstallNamespace,
			resourceutil.WithDeploymentSpec(depSpec.Spec),
			resourceutil.WithLabels(depSpec.Label),
		)

		secretInfo := CertProvisionerFor(depSpec.Name, opts).GetCertSecretInfo()
		if webhookDeployments.Has(depSpec.Name) && secretInfo != nil {
			ensureCorrectDeploymentCertVolumes(deploymentResource, *secretInfo)
		}

		applyCustomConfigToDeployment(deploymentResource, opts.DeploymentConfig)

		objs = append(objs, deploymentResource)
	}
	return objs, nil
}

func bundleCSVPermissionsGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}

	if len(opts.TargetNamespaces) == 1 && opts.TargetNamespaces[0] == "" {
		return nil, nil
	}

	permissions := rv1.CSV.Spec.InstallStrategy.StrategySpec.Permissions

	objs := make([]client.Object, 0, 2*len(opts.TargetNamespaces)*len(permissions))
	for _, ns := range opts.TargetNamespaces {
		for _, permission := range permissions {
			saName := saNameOrDefault(permission.ServiceAccountName)
			name := opts.UniqueNameGenerator(fmt.Sprintf("%s-%s", rv1.CSV.Name, saName), permission)

			objs = append(objs,
				resourceutil.CreateRoleResource(name, ns, resourceutil.WithRules(permission.Rules...)),
				resourceutil.CreateRoleBindingResource(
					name,
					ns,
					resourceutil.WithSubjects(rbacv1.Subject{Kind: "ServiceAccount", Namespace: opts.InstallNamespace, Name: saName}),
					resourceutil.WithRoleRef(rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: name}),
				),
			)
		}
	}
	return objs, nil
}

func bundleCSVClusterPermissionsGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}
	clusterPermissions := rv1.CSV.Spec.InstallStrategy.StrategySpec.ClusterPermissions

	if len(opts.TargetNamespaces) == 1 && opts.TargetNamespaces[0] == "" {
		for _, p := range rv1.CSV.Spec.InstallStrategy.StrategySpec.Permissions {
			p.Rules = append(p.Rules, rbacv1.PolicyRule{
				Verbs:     []string{"get", "list", "watch"},
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"namespaces"},
			})
			clusterPermissions = append(clusterPermissions, p)
		}
	}

	objs := make([]client.Object, 0, 2*len(clusterPermissions))
	for _, permission := range clusterPermissions {
		saName := saNameOrDefault(permission.ServiceAccountName)
		name := opts.UniqueNameGenerator(fmt.Sprintf("%s-%s", rv1.CSV.Name, saName), permission)
		objs = append(objs,
			resourceutil.CreateClusterRoleResource(name, resourceutil.WithRules(permission.Rules...)),
			resourceutil.CreateClusterRoleBindingResource(
				name,
				resourceutil.WithSubjects(rbacv1.Subject{Kind: "ServiceAccount", Namespace: opts.InstallNamespace, Name: saName}),
				resourceutil.WithRoleRef(rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: name}),
			),
		)
	}
	return objs, nil
}

func bundleCSVServiceAccountGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}
	allPermissions := append(
		rv1.CSV.Spec.InstallStrategy.StrategySpec.Permissions,
		rv1.CSV.Spec.InstallStrategy.StrategySpec.ClusterPermissions...,
	)

	serviceAccountNames := sets.Set[string]{}
	for _, permission := range allPermissions {
		serviceAccountNames.Insert(saNameOrDefault(permission.ServiceAccountName))
	}

	objs := make([]client.Object, 0, len(serviceAccountNames))
	for _, serviceAccountName := range serviceAccountNames.UnsortedList() {
		if serviceAccountName != "default" {
			objs = append(objs, resourceutil.CreateServiceAccountResource(serviceAccountName, opts.InstallNamespace))
		}
	}
	return objs, nil
}

func bundleCRDGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}

	crdToDeploymentMap := map[string]v1alpha1.WebhookDescription{}
	for _, wh := range rv1.CSV.Spec.WebhookDefinitions {
		if wh.Type != v1alpha1.ConversionWebhook {
			continue
		}
		for _, crdName := range wh.ConversionCRDs {
			if _, ok := crdToDeploymentMap[crdName]; ok {
				return nil, fmt.Errorf("custom resource definition '%s' is referenced by multiple conversion webhook definitions", crdName)
			}
			crdToDeploymentMap[crdName] = wh
		}
	}

	objs := make([]client.Object, 0, len(rv1.CRDs))
	for _, crd := range rv1.CRDs {
		cp := crd.DeepCopy()
		if cw, ok := crdToDeploymentMap[crd.Name]; ok {
			if crd.Spec.PreserveUnknownFields {
				return nil, fmt.Errorf("custom resource definition '%s' must have .spec.preserveUnknownFields set to false to let API Server call webhook to do the conversion", crd.Name)
			}

			conversionWebhookPath := "/"
			if cw.WebhookPath != nil {
				conversionWebhookPath = *cw.WebhookPath
			}

			certProvisioner := CertProvisionerFor(cw.DeploymentName, opts)
			cp.Spec.Conversion = &apiextensionsv1.CustomResourceConversion{
				Strategy: apiextensionsv1.WebhookConverter,
				Webhook: &apiextensionsv1.WebhookConversion{
					ClientConfig: &apiextensionsv1.WebhookClientConfig{
						Service: &apiextensionsv1.ServiceReference{
							Namespace: opts.InstallNamespace,
							Name:      certProvisioner.ServiceName,
							Path:      &conversionWebhookPath,
							Port:      &cw.ContainerPort,
						},
					},
					ConversionReviewVersions: cw.AdmissionReviewVersions,
				},
			}

			if err := certProvisioner.InjectCABundle(cp); err != nil {
				return nil, err
			}
		}
		objs = append(objs, cp)
	}
	return objs, nil
}

func bundleAdditionalResourcesGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}
	objs := make([]client.Object, 0, len(rv1.Others))
	for _, res := range rv1.Others {
		supported, namespaced := registrybundle.IsSupported(res.GetKind())
		if !supported {
			return nil, fmt.Errorf("bundle contains unsupported resource: Name: %v, Kind: %v", res.GetName(), res.GetKind())
		}

		obj := res.DeepCopy()
		if namespaced {
			obj.SetNamespace(opts.InstallNamespace)
		}

		objs = append(objs, obj)
	}
	return objs, nil
}

func bundleValidatingWebhookResourceGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}

	//nolint:prealloc
	var objs []client.Object

	for _, wh := range rv1.CSV.Spec.WebhookDefinitions {
		if wh.Type != v1alpha1.ValidatingAdmissionWebhook {
			continue
		}
		certProvisioner := CertProvisionerFor(wh.DeploymentName, opts)
		webhookName := strings.TrimSuffix(wh.GenerateName, "-")
		webhookResource := resourceutil.CreateValidatingWebhookConfigurationResource(
			webhookName,
			opts.InstallNamespace,
			resourceutil.WithValidatingWebhooks(
				admissionregistrationv1.ValidatingWebhook{
					Name:                    webhookName,
					Rules:                   wh.Rules,
					FailurePolicy:           wh.FailurePolicy,
					MatchPolicy:             wh.MatchPolicy,
					ObjectSelector:          wh.ObjectSelector,
					SideEffects:             wh.SideEffects,
					TimeoutSeconds:          wh.TimeoutSeconds,
					AdmissionReviewVersions: wh.AdmissionReviewVersions,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: opts.InstallNamespace,
							Name:      certProvisioner.ServiceName,
							Path:      wh.WebhookPath,
							Port:      &wh.ContainerPort,
						},
					},
					NamespaceSelector: getWebhookNamespaceSelector(opts.TargetNamespaces),
				},
			),
		)
		if err := certProvisioner.InjectCABundle(webhookResource); err != nil {
			return nil, err
		}
		objs = append(objs, webhookResource)
	}
	return objs, nil
}

func bundleMutatingWebhookResourceGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}

	//nolint:prealloc
	var objs []client.Object
	for _, wh := range rv1.CSV.Spec.WebhookDefinitions {
		if wh.Type != v1alpha1.MutatingAdmissionWebhook {
			continue
		}
		certProvisioner := CertProvisionerFor(wh.DeploymentName, opts)
		webhookName := strings.TrimSuffix(wh.GenerateName, "-")
		webhookResource := resourceutil.CreateMutatingWebhookConfigurationResource(
			webhookName,
			opts.InstallNamespace,
			resourceutil.WithMutatingWebhooks(
				admissionregistrationv1.MutatingWebhook{
					Name:                    webhookName,
					Rules:                   wh.Rules,
					FailurePolicy:           wh.FailurePolicy,
					MatchPolicy:             wh.MatchPolicy,
					ObjectSelector:          wh.ObjectSelector,
					SideEffects:             wh.SideEffects,
					TimeoutSeconds:          wh.TimeoutSeconds,
					AdmissionReviewVersions: wh.AdmissionReviewVersions,
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: opts.InstallNamespace,
							Name:      certProvisioner.ServiceName,
							Path:      wh.WebhookPath,
							Port:      &wh.ContainerPort,
						},
					},
					ReinvocationPolicy: wh.ReinvocationPolicy,
					NamespaceSelector:  getWebhookNamespaceSelector(opts.TargetNamespaces),
				},
			),
		)
		if err := certProvisioner.InjectCABundle(webhookResource); err != nil {
			return nil, err
		}
		objs = append(objs, webhookResource)
	}
	return objs, nil
}

func bundleDeploymentServiceResourceGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if rv1 == nil {
		return nil, fmt.Errorf("bundle cannot be nil")
	}

	webhookServicePortsByDeployment := map[string]sets.Set[corev1.ServicePort]{}
	for _, wh := range rv1.CSV.Spec.WebhookDefinitions {
		if _, ok := webhookServicePortsByDeployment[wh.DeploymentName]; !ok {
			webhookServicePortsByDeployment[wh.DeploymentName] = sets.Set[corev1.ServicePort]{}
		}
		webhookServicePortsByDeployment[wh.DeploymentName].Insert(getWebhookServicePort(wh))
	}

	objs := make([]client.Object, 0, len(webhookServicePortsByDeployment))
	for _, deploymentSpec := range rv1.CSV.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		if _, ok := webhookServicePortsByDeployment[deploymentSpec.Name]; !ok {
			continue
		}

		servicePorts := webhookServicePortsByDeployment[deploymentSpec.Name]
		ports := servicePorts.UnsortedList()
		slices.SortStableFunc(ports, func(a, b corev1.ServicePort) int {
			return cmp.Or(cmp.Compare(a.Port, b.Port), cmp.Compare(a.TargetPort.IntValue(), b.TargetPort.IntValue()))
		})

		var labelSelector map[string]string
		if deploymentSpec.Spec.Selector != nil {
			labelSelector = deploymentSpec.Spec.Selector.MatchLabels
		}

		certProvisioner := CertProvisionerFor(deploymentSpec.Name, opts)
		serviceResource := resourceutil.CreateServiceResource(
			certProvisioner.ServiceName,
			opts.InstallNamespace,
			resourceutil.WithServiceSpec(
				corev1.ServiceSpec{
					Ports:    ports,
					Selector: labelSelector,
				},
			),
		)

		if err := certProvisioner.InjectCABundle(serviceResource); err != nil {
			return nil, err
		}
		objs = append(objs, serviceResource)
	}

	return objs, nil
}

func certProviderResourceGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	deploymentsWithWebhooks := sets.Set[string]{}

	for _, wh := range rv1.CSV.Spec.WebhookDefinitions {
		deploymentsWithWebhooks.Insert(wh.DeploymentName)
	}

	var objs []client.Object
	for _, depName := range deploymentsWithWebhooks.UnsortedList() {
		certCfg := CertProvisionerFor(depName, opts)
		certObjs, err := certCfg.AdditionalObjects()
		if err != nil {
			return nil, err
		}
		for _, certObj := range certObjs {
			objs = append(objs, &certObj)
		}
	}
	return objs, nil
}

func saNameOrDefault(saName string) string {
	return cmp.Or(saName, "default")
}

func getWebhookServicePort(wh v1alpha1.WebhookDescription) corev1.ServicePort {
	containerPort := int32(443)
	if wh.ContainerPort > 0 {
		containerPort = wh.ContainerPort
	}

	targetPort := intstr.FromInt32(containerPort)
	if wh.TargetPort != nil {
		targetPort = *wh.TargetPort
	}

	return corev1.ServicePort{
		Name:       strconv.Itoa(int(containerPort)),
		Port:       containerPort,
		TargetPort: targetPort,
	}
}

func ensureCorrectDeploymentCertVolumes(dep *appsv1.Deployment, certSecretInfo CertSecretInfo) {
	volumesToRemove := sets.New[string]()
	protectedVolumePaths := sets.New[string]()
	certVolumes := make([]corev1.Volume, 0, len(certVolumeConfigs))
	certVolumeMounts := make([]corev1.VolumeMount, 0, len(certVolumeConfigs))
	for _, cfg := range certVolumeConfigs {
		volumesToRemove.Insert(cfg.Name)
		protectedVolumePaths.Insert(cfg.Path)
		certVolumes = append(certVolumes, corev1.Volume{
			Name: cfg.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: certSecretInfo.SecretName,
					Items: []corev1.KeyToPath{
						{
							Key:  certSecretInfo.CertificateKey,
							Path: cfg.TLSCertPath,
						},
						{
							Key:  certSecretInfo.PrivateKeyKey,
							Path: cfg.TLSKeyPath,
						},
					},
				},
			},
		})
		certVolumeMounts = append(certVolumeMounts, corev1.VolumeMount{
			Name:      cfg.Name,
			MountPath: cfg.Path,
		})
	}

	for _, c := range dep.Spec.Template.Spec.Containers {
		for _, containerVolumeMount := range c.VolumeMounts {
			if protectedVolumePaths.Has(containerVolumeMount.MountPath) {
				volumesToRemove.Insert(containerVolumeMount.Name)
			}
		}
	}

	dep.Spec.Template.Spec.Volumes = slices.Concat(
		slices.DeleteFunc(dep.Spec.Template.Spec.Volumes, func(v corev1.Volume) bool {
			return volumesToRemove.Has(v.Name)
		}),
		certVolumes,
	)

	for i := range dep.Spec.Template.Spec.Containers {
		dep.Spec.Template.Spec.Containers[i].VolumeMounts = slices.Concat(
			slices.DeleteFunc(dep.Spec.Template.Spec.Containers[i].VolumeMounts, func(v corev1.VolumeMount) bool {
				return volumesToRemove.Has(v.Name)
			}),
			certVolumeMounts,
		)
	}
}

func getWebhookNamespaceSelector(targetNamespaces []string) *metav1.LabelSelector {
	if len(targetNamespaces) > 0 && !slices.Contains(targetNamespaces, "") {
		return &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      labelKubernetesNamespaceMetadataName,
					Operator: metav1.LabelSelectorOpIn,
					Values:   targetNamespaces,
				},
			},
		}
	}
	return nil
}

func applyCustomConfigToDeployment(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if config == nil {
		return
	}

	applyEnvironmentConfig(deployment, config)
	applyEnvironmentFromConfig(deployment, config)
	applyVolumeConfig(deployment, config)
	applyVolumeMountConfig(deployment, config)
	applyTolerationsConfig(deployment, config)
	applyResourcesConfig(deployment, config)
	applyNodeSelectorConfig(deployment, config)
	applyAffinityConfig(deployment, config)
	applyAnnotationsConfig(deployment, config)
}

func applyEnvironmentConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if len(config.Env) == 0 {
		return
	}

	for i := range deployment.Spec.Template.Spec.Containers {
		container := &deployment.Spec.Template.Spec.Containers[i]

		existingEnvMap := make(map[string]int)
		for idx, env := range container.Env {
			existingEnvMap[env.Name] = idx
		}

		for _, configEnv := range config.Env {
			if existingIdx, exists := existingEnvMap[configEnv.Name]; exists {
				container.Env[existingIdx] = configEnv
			} else {
				container.Env = append(container.Env, configEnv)
			}
		}
	}
}

func applyEnvironmentFromConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if len(config.EnvFrom) == 0 {
		return
	}

	for i := range deployment.Spec.Template.Spec.Containers {
		container := &deployment.Spec.Template.Spec.Containers[i]

		for _, configEnvFrom := range config.EnvFrom {
			isDuplicate := false
			for _, existingEnvFrom := range container.EnvFrom {
				if reflect.DeepEqual(existingEnvFrom, configEnvFrom) {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				container.EnvFrom = append(container.EnvFrom, configEnvFrom)
			}
		}
	}
}

func applyVolumeConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if len(config.Volumes) == 0 {
		return
	}

	existingVolMap := make(map[string]int, len(deployment.Spec.Template.Spec.Volumes))
	for i, vol := range deployment.Spec.Template.Spec.Volumes {
		existingVolMap[vol.Name] = i
	}

	for _, configVol := range config.Volumes {
		if idx, exists := existingVolMap[configVol.Name]; exists {
			deployment.Spec.Template.Spec.Volumes[idx] = configVol
		} else {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, configVol)
		}
	}
}

func applyVolumeMountConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if len(config.VolumeMounts) == 0 {
		return
	}

	for i := range deployment.Spec.Template.Spec.Containers {
		container := &deployment.Spec.Template.Spec.Containers[i]

		existingMountMap := make(map[string]int, len(container.VolumeMounts))
		for idx, mount := range container.VolumeMounts {
			existingMountMap[mount.Name] = idx
		}

		for _, configMount := range config.VolumeMounts {
			if idx, exists := existingMountMap[configMount.Name]; exists {
				container.VolumeMounts[idx] = configMount
			} else {
				container.VolumeMounts = append(container.VolumeMounts, configMount)
			}
		}
	}
}

func applyTolerationsConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if len(config.Tolerations) == 0 {
		return
	}

	for _, configToleration := range config.Tolerations {
		isDuplicate := false
		for _, existingToleration := range deployment.Spec.Template.Spec.Tolerations {
			if reflect.DeepEqual(existingToleration, configToleration) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, configToleration)
		}
	}
}

func applyResourcesConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if config.Resources == nil {
		return
	}

	for i := range deployment.Spec.Template.Spec.Containers {
		container := &deployment.Spec.Template.Spec.Containers[i]
		container.Resources = *config.Resources
	}
}

func applyNodeSelectorConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if config.NodeSelector == nil {
		return
	}

	deployment.Spec.Template.Spec.NodeSelector = config.NodeSelector
}

func isAffinityEmpty(a *corev1.Affinity) bool {
	if a == nil {
		return true
	}
	return isNodeAffinityEmpty(a.NodeAffinity) &&
		isPodAffinityEmpty(a.PodAffinity) &&
		isPodAntiAffinityEmpty(a.PodAntiAffinity)
}

func isNodeAffinityEmpty(na *corev1.NodeAffinity) bool {
	if na == nil {
		return true
	}
	requiredEmpty := na.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		len(na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0
	return requiredEmpty && len(na.PreferredDuringSchedulingIgnoredDuringExecution) == 0
}

func isPodAffinityEmpty(pa *corev1.PodAffinity) bool {
	if pa == nil {
		return true
	}
	return len(pa.RequiredDuringSchedulingIgnoredDuringExecution) == 0 &&
		len(pa.PreferredDuringSchedulingIgnoredDuringExecution) == 0
}

func isPodAntiAffinityEmpty(paa *corev1.PodAntiAffinity) bool {
	if paa == nil {
		return true
	}
	return len(paa.RequiredDuringSchedulingIgnoredDuringExecution) == 0 &&
		len(paa.PreferredDuringSchedulingIgnoredDuringExecution) == 0
}

func applyAffinityConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if config.Affinity == nil {
		return
	}

	podSpec := &deployment.Spec.Template.Spec

	configHasNoFields := config.Affinity.NodeAffinity == nil &&
		config.Affinity.PodAffinity == nil &&
		config.Affinity.PodAntiAffinity == nil

	if configHasNoFields {
		podSpec.Affinity = nil
		return
	}

	if podSpec.Affinity == nil {
		podSpec.Affinity = &corev1.Affinity{}
	}

	if config.Affinity.NodeAffinity != nil {
		if isNodeAffinityEmpty(config.Affinity.NodeAffinity) {
			podSpec.Affinity.NodeAffinity = nil
		} else {
			podSpec.Affinity.NodeAffinity = config.Affinity.NodeAffinity
		}
	}

	if config.Affinity.PodAffinity != nil {
		if isPodAffinityEmpty(config.Affinity.PodAffinity) {
			podSpec.Affinity.PodAffinity = nil
		} else {
			podSpec.Affinity.PodAffinity = config.Affinity.PodAffinity
		}
	}

	if config.Affinity.PodAntiAffinity != nil {
		if isPodAntiAffinityEmpty(config.Affinity.PodAntiAffinity) {
			podSpec.Affinity.PodAntiAffinity = nil
		} else {
			podSpec.Affinity.PodAntiAffinity = config.Affinity.PodAntiAffinity
		}
	}

	if isAffinityEmpty(podSpec.Affinity) {
		podSpec.Affinity = nil
	}
}

func applyAnnotationsConfig(deployment *appsv1.Deployment, config *DeploymentConfig) {
	if len(config.Annotations) == 0 {
		return
	}

	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	for key, value := range config.Annotations {
		if _, exists := deployment.Annotations[key]; !exists {
			deployment.Annotations[key] = value
		}
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	for key, value := range config.Annotations {
		if _, exists := deployment.Spec.Template.Annotations[key]; !exists {
			deployment.Spec.Template.Annotations[key] = value
		}
	}
}

var (
	adminVerbs = []string{"*"}
	editVerbs  = []string{"create", "update", "patch", "delete"}
	viewVerbs  = []string{"get", "list", "watch"}

	verbsForSuffix = map[string][]string{
		"admin": adminVerbs,
		"edit":  editVerbs,
		"view":  viewVerbs,
	}
)

func bundleProvidedAPIsClusterRolesGenerator(rv1 *bundle.RegistryV1, opts options) ([]client.Object, error) {
	if !opts.ProvidedAPIsClusterRoles {
		return nil, nil
	}
	objects := make([]client.Object, 0, len(rv1.CSV.Spec.CustomResourceDefinitions.Owned)*4)

	for _, owned := range rv1.CSV.Spec.CustomResourceDefinitions.Owned {
		nameGroupPair := strings.SplitN(owned.Name, ".", 2)
		if len(nameGroupPair) != 2 {
			return nil, fmt.Errorf("invalid CRD name %q: expected <plural>.<group>", owned.Name)
		}
		plural := nameGroupPair[0]
		group := nameGroupPair[1]
		namePrefix := fmt.Sprintf("%s-%s-", owned.Name, owned.Version)

		for suffix, verbs := range verbsForSuffix {
			objects = append(objects, &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: namePrefix + suffix,
					Labels: map[string]string{
						fmt.Sprintf("rbac.authorization.k8s.io/aggregate-to-%s", suffix): "true",
					},
				},
				Rules: []rbacv1.PolicyRule{{
					Verbs:     verbs,
					APIGroups: []string{group},
					Resources: []string{plural},
				}},
			})
		}

		objects = append(objects, &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix + "crd-view",
				Labels: map[string]string{
					"rbac.authorization.k8s.io/aggregate-to-view": "true",
				},
			},
			Rules: []rbacv1.PolicyRule{{
				Verbs:         []string{"get"},
				APIGroups:     []string{"apiextensions.k8s.io"},
				Resources:     []string{"customresourcedefinitions"},
				ResourceNames: []string{owned.Name},
			}},
		})
	}

	return objects, nil
}

// defaultGenerators returns the standard set of resource generators for registry+v1 bundles.
func defaultGenerators() []resourceGenerator {
	return []resourceGenerator{
		bundleCSVServiceAccountGenerator,
		bundleCSVPermissionsGenerator,
		bundleCSVClusterPermissionsGenerator,
		bundleCRDGenerator,
		bundleAdditionalResourcesGenerator,
		bundleCSVDeploymentGenerator,
		bundleValidatingWebhookResourceGenerator,
		bundleMutatingWebhookResourceGenerator,
		bundleDeploymentServiceResourceGenerator,
		certProviderResourceGenerator,
		bundleProvidedAPIsClusterRolesGenerator,
	}
}
