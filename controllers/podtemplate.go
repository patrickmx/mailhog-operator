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

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("200m")
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("150Mi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("200m")
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("150Mi")

	socketProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(1025),
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
				Port:   intstr.FromInt(8025),
				Path:   "/api/v2/messages?limit=1",
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
		const (
			storageVolumeName  = "maildir-storage"
			settingsVolumeName = "settings-files"
		)
		podVolumes := make([]corev1.Volume, 0)
		containerVolMounts := make([]corev1.VolumeMount, 0)
		if cr.Spec.Settings.StorageMaildir.Path != "" && cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {

			podVolumes = append(podVolumes, corev1.Volume{
				Name: storageVolumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      storageVolumeName,
				MountPath: cr.Spec.Settings.StorageMaildir.Path,
			})

		}
		if cr.Spec.Settings.Files != nil {
			podVolumes = append(podVolumes, corev1.Volume{
				Name: settingsVolumeName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name,
						},
					},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      settingsVolumeName,
				MountPath: settingsFilesMount,
			})

		}
		pod.Spec.Volumes = podVolumes
		pod.Spec.Containers[0].VolumeMounts = containerVolMounts
	}

	return pod
}
