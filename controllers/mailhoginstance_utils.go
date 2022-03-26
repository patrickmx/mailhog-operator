package controllers

import (
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func portsForCr() (p []corev1.ContainerPort) {
	return []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: 8025,
			Protocol:      "TCP",
		},
		{
			Name:          "smtp",
			ContainerPort: 1025,
			Protocol:      "TCP",
		},
	}
}

func envForCr(crs *mailhogv1alpha1.MailhogInstance) (e []corev1.EnvVar) {
	e = append(e, corev1.EnvVar{
		Name:  "MH_SMTP_BIND_ADDR",
		Value: "0.0.0.0:1025",
	})

	e = append(e, corev1.EnvVar{
		Name:  "MH_API_BIND_ADDR",
		Value: "0.0.0.0:8025",
	})

	e = append(e, corev1.EnvVar{
		Name:  "MH_UI_BIND_ADDR",
		Value: "0.0.0.0:8025",
	})

	if crs.Spec.Settings.Storage != "" {
		e = append(e, corev1.EnvVar{
			Name:  "MH_STORAGE",
			Value: string(crs.Spec.Settings.Storage),
		})
	}

	if crs.Spec.Settings.Storage == mailhogv1alpha1.MongoDBStorage {
		if crs.Spec.Settings.StorageMongoDb.URI != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MONGO_URI",
				Value: crs.Spec.Settings.StorageMongoDb.URI,
			})
		}

		if crs.Spec.Settings.StorageMongoDb.Db != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MONGO_DB",
				Value: crs.Spec.Settings.StorageMongoDb.Db,
			})
		}

		if crs.Spec.Settings.StorageMongoDb.Collection != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MONGO_COLLECTION",
				Value: crs.Spec.Settings.StorageMongoDb.Collection,
			})
		}
	}

	if crs.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {
		if crs.Spec.Settings.StorageMaildir.Path != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MAILDIR_PATH",
				Value: crs.Spec.Settings.StorageMaildir.Path,
			})
		}
	}

	if crs.Spec.Settings.Hostname != "" {
		e = append(e, corev1.EnvVar{
			Name:  "MH_HOSTNAME",
			Value: crs.Spec.Settings.Hostname,
		})
	}

	if crs.Spec.Settings.CorsOrigin != "" {
		e = append(e, corev1.EnvVar{
			Name:  "MH_CORS_ORIGIN",
			Value: crs.Spec.Settings.CorsOrigin,
		})
	}

	return
}

func labelsForCr(name string) map[string]string {
	return map[string]string{"app": "mailhog", "mailhog_cr": name}
}

func textLabelsForCr(name string) string {
	return "app=mailhog,mailhog_cr=" + name
}

func getPodNames(pods []corev1.Pod) []string {
	podNames := make([]string, 0)
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
