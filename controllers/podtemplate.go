package controllers

import (
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *MailhogInstanceReconciler) podTemplate(cr *mailhogv1alpha1.MailhogInstance) corev1.PodTemplateSpec {
	labels := labelsForCr(cr.Name)
	labels["deploymentconfig"] = cr.Name
	env := envForCr(cr)
	ports := portsForCr()
	image := cr.Spec.Image

	var resources corev1.ResourceRequirements
	if cr.Spec.Settings.Resources == nil {
		resources = defaultResources()
	} else {
		resources = *cr.Spec.Settings.Resources.DeepCopy()
	}

	socketProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(portWeb),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	httpProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromInt(portWeb),
				Path:   cr.Spec.Settings.WebPath + "/api/v2/messages?limit=1",
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	isExplicitlyFalse := false

	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:           "mailhog",
					Image:          image,
					Ports:          ports,
					Env:            env,
					Resources:      resources,
					LivenessProbe:  socketProbe,
					StartupProbe:   socketProbe,
					ReadinessProbe: httpProbe,
				},
			},
			AutomountServiceAccountToken: &isExplicitlyFalse,
		},
	}

	if cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage || cr.Spec.Settings.Files != nil {
		podVolumes := make([]corev1.Volume, 0)
		containerVolMounts := make([]corev1.VolumeMount, 0)
		if cr.Spec.Settings.StorageMaildir.Path != "" && cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {

			podVolumes = append(podVolumes, corev1.Volume{
				Name: volumeNameMaildir,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      volumeNameMaildir,
				MountPath: cr.Spec.Settings.StorageMaildir.Path,
			})

		}
		if cr.Spec.Settings.Files != nil {
			podVolumes = append(podVolumes, corev1.Volume{
				Name: volumeNameSettings,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name,
						},
					},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      volumeNameSettings,
				MountPath: settingsFilesMount,
			})

		}
		pod.Spec.Volumes = podVolumes
		pod.Spec.Containers[0].VolumeMounts = containerVolMounts
	}

	if cr.Spec.Settings.Jim.Invite == true {
		pod.Spec.Containers[0].Args = jimArgs(cr)
	}

	if cr.Spec.Settings.Affinity != nil {
		pod.Spec.Affinity = cr.Spec.Settings.Affinity.DeepCopy()
	}

	if cr.Spec.Settings.Files != nil && len(cr.Spec.Settings.Files.WebUsers) > 0 {
		// since http authentication is active, kube can no longer perform a http health check, switch to socket
		pod.Spec.Containers[0].ReadinessProbe = socketProbe.DeepCopy()
	}

	return pod
}

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

func jimArgs(cr *mailhogv1alpha1.MailhogInstance) []string {
	args := make([]string, 0)

	if cr.Spec.Settings.Jim.Invite == true {
		args = append(args, "-invite-jim")
		args = appendNonEmptyArg(args, "jim-disconnect", cr.Spec.Settings.Jim.Disconnect)
		args = appendNonEmptyArg(args, "jim-accpet", cr.Spec.Settings.Jim.Accept)
		args = appendNonEmptyArg(args, "jim-linkspeed-affect", cr.Spec.Settings.Jim.LinkspeedAffect)
		args = appendNonEmptyArg(args, "jim-linkspeed-min", cr.Spec.Settings.Jim.LinkspeedMin)
		args = appendNonEmptyArg(args, "jim-linkspeed-max", cr.Spec.Settings.Jim.LinkspeedMax)
		args = appendNonEmptyArg(args, "jim-reject-sender", cr.Spec.Settings.Jim.RejectSender)
		args = appendNonEmptyArg(args, "jim-reject-recipient", cr.Spec.Settings.Jim.RejectRecipient)
		args = appendNonEmptyArg(args, "jim-reject-auth", cr.Spec.Settings.Jim.RejectAuth)
	}

	return args
}

func appendNonEmptyArg(args []string, arg string, value string) []string {
	if value == "" {
		return args
	}
	args = append(args, "-"+arg+"="+value)
	return args
}
