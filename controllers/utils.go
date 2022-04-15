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
			Protocol:      protoTcp,
		},
		{
			Name:          portSmtpName,
			ContainerPort: portSmtp,
			Protocol:      protoTcp,
		},
	}
}

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

func getPodNames(pods []corev1.Pod) []string {
	podNames := make([]string, 0)
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
