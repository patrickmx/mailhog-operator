package controllers

import (
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// podTemplate will return the desired PodTemplate used in Deployments and DeploymentConfigs
func podTemplate(cr *mailhogv1alpha1.MailhogInstance) corev1.PodTemplateSpec {
	meta := CreateMetaMaker(cr)
	env := envForCr(cr)
	ports := portsForCr()
	image := cr.Spec.Image

	var resources corev1.ResourceRequirements
	if cr.Spec.Settings.Resources == nil {
		resources = defaultResources()
	} else {
		resources = *cr.Spec.Settings.Resources.DeepCopy()
	}

	isExplicitlyFalse := false

	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: meta.GetLabels(true),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:           mh,
					Image:          image,
					Ports:          ports,
					Env:            env,
					Resources:      resources,
					LivenessProbe:  getProbeTcp(portWeb),
					StartupProbe:   getProbeTcp(portWeb),
					ReadinessProbe: getProbeHttp(portWeb, cr.Spec.Settings.WebPath+httpHealthPath),
				},
			},
			AutomountServiceAccountToken: &isExplicitlyFalse,
		},
	}

	if cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage || cr.Spec.Settings.Files != nil {
		pod.Spec.Volumes, pod.Spec.Containers[0].VolumeMounts = podVolumes(cr)
	}

	if cr.Spec.Settings.Jim.Invite == true {
		pod.Spec.Containers[0].Args = jimArgs(cr)
	}

	if cr.Spec.Settings.Affinity != nil {
		pod.Spec.Affinity = cr.Spec.Settings.Affinity.DeepCopy()
	}

	if cr.Spec.Settings.Files != nil && len(cr.Spec.Settings.Files.WebUsers) > 0 {
		// since http authentication is active, kube can no longer perform a http health check, switch to socket
		pod.Spec.Containers[0].ReadinessProbe = getProbeTcp(portWeb)
	}

	return pod
}

// podVolumes will return the required volumes and mounts for a given CR
func podVolumes(cr *mailhogv1alpha1.MailhogInstance) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) {
	if cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage || cr.Spec.Settings.Files != nil {
		if cr.Spec.Settings.StorageMaildir.Path != "" && cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {

			if claimName := cr.Spec.Settings.StorageMaildir.PvName; claimName == "" {
				volumes = append(volumes, corev1.Volume{
					Name: volumeNameMaildir,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				})
			} else {
				volumes = append(volumes, corev1.Volume{
					Name: volumeNameMaildir,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimName,
							ReadOnly:  false,
						},
					},
				})
			}
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      volumeNameMaildir,
				MountPath: cr.Spec.Settings.StorageMaildir.Path,
			})

		}
		if cr.Spec.Settings.Files != nil {
			volumes = append(volumes, corev1.Volume{
				Name: volumeNameSettings,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name,
						},
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      volumeNameSettings,
				MountPath: settingsFilesMount,
			})

		}
	}
	return volumes, volumeMounts
}

// defaultResources will return the resource limits used when none are specified in the CR
func defaultResources() corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	resources.Requests[corev1.ResourceCPU] = resource.MustParse(defaultResourceCPU)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse(defaultResourceMemory)
	resources.Limits[corev1.ResourceCPU] = resource.MustParse(defaultResourceCPU)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse(defaultResourceMemory)
	return resources
}

// getProbeHttp will return a new http probe on the given port and path
func getProbeHttp(port int, path string) (probe *corev1.Probe) {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromInt(port),
				Path:   path,
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// getProbeTcp will return a new tcp probe on the given port
func getProbeTcp(port int) (probe *corev1.Probe) {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(port),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// jimArgs will return the needed container args for the desired chaos monkey configuration
func jimArgs(cr *mailhogv1alpha1.MailhogInstance) []string {
	args := make([]string, 0)

	if cr.Spec.Settings.Jim.Invite == true {
		args = append(args, argsJimInvite)
		args = appendNonEmptyArg(args, argsJimDisconnectRate, cr.Spec.Settings.Jim.Disconnect)
		args = appendNonEmptyArg(args, argsJimAccept, cr.Spec.Settings.Jim.Accept)
		args = appendNonEmptyArg(args, argsJimLinkSpeedAffect, cr.Spec.Settings.Jim.LinkspeedAffect)
		args = appendNonEmptyArg(args, argsJimLinkSpeedMin, cr.Spec.Settings.Jim.LinkspeedMin)
		args = appendNonEmptyArg(args, argsJimLinkSpeedMax, cr.Spec.Settings.Jim.LinkspeedMax)
		args = appendNonEmptyArg(args, argsJimRejectSender, cr.Spec.Settings.Jim.RejectSender)
		args = appendNonEmptyArg(args, argsJimRejectRecipient, cr.Spec.Settings.Jim.RejectRecipient)
		args = appendNonEmptyArg(args, argsJimRejectAuth, cr.Spec.Settings.Jim.RejectAuth)
	}

	return args
}

// appendNonemptyArgs will append container args if the value is non-empty
func appendNonEmptyArg(args []string, arg string, value string) []string {
	if value == "" {
		return args
	}
	args = append(args, "-"+arg+"="+value)
	return args
}

// portsForCr will return the desired ContainerPorts of a given CR
func portsForCr() (p []corev1.ContainerPort) {
	return []corev1.ContainerPort{
		{
			Name:          portWebName,
			ContainerPort: portWeb,
			Protocol:      protoTcp,
		},
		{
			Name:          portSmtpName,
			ContainerPort: portSmtp,
			Protocol:      protoTcp,
		},
	}
}

// envForCr will return the desired environment variables for a give CR
func envForCr(crs *mailhogv1alpha1.MailhogInstance) (e []corev1.EnvVar) {
	e = []corev1.EnvVar{
		{
			Name:  envSmtpBind,
			Value: envBindSmtpValue,
		},
		{
			Name:  envApiBind,
			Value: envBindWebValue,
		},
		{
			Name:  envUiBind,
			Value: envBindWebValue,
		},
	}

	e = appendNonEmptyEnv(e, envStorage, string(crs.Spec.Settings.Storage))

	if crs.Spec.Settings.Storage == mailhogv1alpha1.MongoDBStorage {
		e = appendNonEmptyEnv(e, envMongoUri, crs.Spec.Settings.StorageMongoDb.URI)
		e = appendNonEmptyEnv(e, envMongoDb, crs.Spec.Settings.StorageMongoDb.Db)
		e = appendNonEmptyEnv(e, envMongoCollection, crs.Spec.Settings.StorageMongoDb.Collection)
	}

	if crs.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {
		e = appendNonEmptyEnv(e, envMaildirPath, crs.Spec.Settings.StorageMaildir.Path)
	}

	e = appendNonEmptyEnv(e, envHostname, crs.Spec.Settings.Hostname)
	e = appendNonEmptyEnv(e, envCorsOrigin, crs.Spec.Settings.CorsOrigin)
	e = appendNonEmptyEnv(e, envWebPath, crs.Spec.Settings.WebPath)

	if crs.Spec.Settings.Files != nil {
		if len(crs.Spec.Settings.Files.SmtpUpstreams) > 0 {
			e = appendNonEmptyEnv(e, envUpstreamSmtpFile, settingsFileUpstreamsPath)
		}

		if len(crs.Spec.Settings.Files.WebUsers) > 0 {
			e = appendNonEmptyEnv(e, envWebAuthFile, settingsFilePasswordsPath)
		}
	}

	return
}

// appendNonEmptyEnv will append to the environment if the value is non-empty
func appendNonEmptyEnv(env []corev1.EnvVar, key string, value string) []corev1.EnvVar {
	if value == "" {
		return env
	}
	env = append(env, corev1.EnvVar{
		Name:  key,
		Value: value,
	})
	return env
}
