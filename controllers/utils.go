package controllers

import (
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type (
	ReturnIndicator struct {
		Err error
	}
)

func portsForCr() (p []corev1.ContainerPort) {
	return []corev1.ContainerPort{
		{
			Name:          portWebName,
			ContainerPort: portWeb,
			Protocol:      "TCP",
		},
		{
			Name:          portSmtpName,
			ContainerPort: portSmtp,
			Protocol:      "TCP",
		},
	}
}

func envForCr(crs *mailhogv1alpha1.MailhogInstance) (e []corev1.EnvVar) {
	e = []corev1.EnvVar{
		{
			Name:  "MH_SMTP_BIND_ADDR",
			Value: "0.0.0.0:1025",
		},
		{
			Name:  "MH_API_BIND_ADDR",
			Value: "0.0.0.0:8025",
		},
		{
			Name:  "MH_UI_BIND_ADDR",
			Value: "0.0.0.0:8025",
		},
	}

	e = appendNonEmptyEnv(e, "MH_STORAGE", string(crs.Spec.Settings.Storage))

	if crs.Spec.Settings.Storage == mailhogv1alpha1.MongoDBStorage {
		e = appendNonEmptyEnv(e, "MH_MONGO_URI", crs.Spec.Settings.StorageMongoDb.URI)
		e = appendNonEmptyEnv(e, "MH_MONGO_DB", crs.Spec.Settings.StorageMongoDb.Db)
		e = appendNonEmptyEnv(e, "MH_MONGO_COLLECTION", crs.Spec.Settings.StorageMongoDb.Collection)
	}

	if crs.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {
		e = appendNonEmptyEnv(e, "MH_MAILDIR_PATH", crs.Spec.Settings.StorageMaildir.Path)
	}

	e = appendNonEmptyEnv(e, "MH_HOSTNAME", crs.Spec.Settings.Hostname)
	e = appendNonEmptyEnv(e, "MH_CORS_ORIGIN", crs.Spec.Settings.CorsOrigin)
	e = appendNonEmptyEnv(e, "MH_UI_WEB_PATH", crs.Spec.Settings.WebPath)

	if crs.Spec.Settings.Files != nil {
		if len(crs.Spec.Settings.Files.SmtpUpstreams) > 0 {
			e = appendNonEmptyEnv(e, "MH_OUTGOING_SMTP", settingsFileUpstreamsPath)
		}

		if len(crs.Spec.Settings.Files.WebUsers) > 0 {
			e = appendNonEmptyEnv(e, "MH_AUTH_FILE", settingsFilePasswordsPath)
		}
	}

	return
}

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

func labelsForCr(name string) map[string]string {
	return map[string]string{
		"app":                        "mailhog",
		"mailhog_cr":                 name,
		"app.openshift.io/runtime":   "golang",
		"app.kubernetes.io/name":     "mailhog",
		"app.kubernetes.io/instance": name,
	}
}

func annotationsForCr() map[string]string {
	return map[string]string{
		"app.openshift.io/vcs-uri": "https://github.com/mailhog/MailHog",
	}
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
