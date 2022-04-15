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

	messageFailedGetInitialObject = "failed to annotate new object with initial state"
	messageFailedSetOwnerRef      = "failed to set controller ref for new object"
	messageFailedCreate           = "failed to create new object"
	messageCreatedObject          = "created new object"

	messageFailedGetDeletingObject = "cant check for to-be-removed object"
	messageFailedDelete            = "cant remove obsolete object"
	messageDeletedObject           = "removed obsolete object"

	messageFailedSetOwnerRefUpdate   = "cant set owner reference of updated object"
	messageFailedDeleteAfterInvalid  = "cant remove object which failed to update"
	messageDeletedObjectAfterInvalid = "deleted object because update failed"
	messageFailedUpdate              = "cant update object"
	messageUpdated                   = "updated existing object"

	eventUpdated = "updated by mailhog management"

	failedGetExisting = "failed to get existing object"
	failedUpdateCheck = "failed to check if object needs an update"
	stateEnsured      = "object state ensured"

	failedListPods       = "failed to list pods"
	failedCrRefresh      = "failed to get latest cr version before update"
	failedCrUpdateStatus = "failed to update cr status"
	updatedCrStatus      = "updated cr status"
	noCrUpdateNeeded     = "no cr status update required"

	span                 = "span"
	spanCr               = "crStatus"
	spanService          = "service"
	spanRoute            = "route"
	spanDeployment       = "deployment"
	spanDeploymentConfig = "deploymentConfig"
	spanConfigMap        = "configMap"

	crGetNotFound = "cr not found, probably it was deleted"
	crGetFailed   = "failed to get cr"

	reconcileStarted  = "staring reconcile"
	reconcileFinished = "reconciliation finished, nothing to do"
)
