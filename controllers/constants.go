package controllers

const (
	lastApplied = "mailhog.operators.patrick.mx/last-applied"

	defaultResourceCPU    = "200m"
	defaultResourceMemory = "150Mi"

	volumeNameMaildir  = "maildir-storage"
	volumeNameSettings = "settings-files"

	settingsFilesMount        = "/mailhog/settings/files"
	settingsFileUpstreamsName = "upstream.servers.json"
	settingsFileUpstreamsPath = settingsFilesMount + "/" + "upstream.servers.json"
	//#nosec G101
	settingsFilePasswordsName = "users.list.bcrypt"
	settingsFilePasswordsPath = settingsFilesMount + "/" + "users.list.bcrypt"

	portSmtp     = 1025
	portSmtpName = "smtp"
	portWeb      = 8025
	portWebName  = "http"

	crNameLabel          = "mailhoginstance_cr"
	crTypeLabel          = "mailhogtype"
	crTypeValue          = "mailhoginstance"
	dcLabel              = "deploymentconfig"
	runtimeLabel         = "app.openshift.io/runtime"
	runtimeDefaultValue  = "golang"
	partOfLabel          = "app.kubernetes.io/part-of"
	connectsToAnnotation = "app.openshift.io/connects-to"
	vcsUriAnnotation     = "app.openshift.io/vcs-uri"
	appLabel             = "app"
	kubeAppLabel         = "app.kubernetes.io/name"
	instanceLabel        = "app.kubernetes.io/instance"
	componentLabel       = "app.kubernetes.io/component"
	managedByLabel       = "app.kubernetes.io/managed-by"
	createdByLabel       = "app.kubernetes.io/created-by"
	operatorValue        = "mailhog.operators.patrick.mx"
)
