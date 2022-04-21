package controllers

const (
	lastApplied = "mailhog.operators.patrick.mx/last-applied"
	mh          = "mailhog"

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
	protoTcp     = "TCP"

	crNameLabel          = "mailhoginstance_cr"
	crTypeLabel          = "mailhogtype"
	crTypeValue          = "mailhoginstance"
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

	messageFailedGetDeletingObject = "failed to check for to-be-removed object"
	messageFailedDelete            = "failed to remove obsolete object"
	messageDeletedObject           = "removed obsolete object"

	messageFailedSetOwnerRefUpdate   = "failed to set owner reference of updated object"
	messageFailedDeleteAfterInvalid  = "failed to delete object after it failed to update"
	messageDeletedObjectAfterInvalid = "deleted object after it failed to update"
	messageFailedUpdate              = "failed to update object"
	messageUpdated                   = "updated existing object"

	eventCreated = "child resource create by mailhog-operator"
	eventUpdated = "child resource updated by mailhog-operator"
	eventDeleted = "child resource deleted by mailhog-operator"

	failedGetExisting = "failed to get existing object"
	failedUpdateCheck = "failed to check if object needs an update"
	stateEnsured      = "object state ensured"

	failedListPods       = "failed to list pods"
	failedListRoutes     = "failed to list routes"
	failedCrRefresh      = "failed to get latest cr version before update"
	failedCrUpdateStatus = "failed to update cr status"
	updatedCrStatus      = "updated cr status"
	noCrUpdateNeeded     = "no cr status update required"

	span           = "span"
	spanCrValid    = "cr.validation"
	spanCrStatus   = "cr.status"
	spanService    = "service"
	spanRoute      = "route"
	spanDeployment = "deployment"
	spanConfigMap  = "configMap"
	spanIgress     = "ingress"

	crGetNotFound = "cr not found, probably it was deleted"
	crGetFailed   = "failed to get cr"

	reconcileStarted  = "staring reconcile"
	reconcileFinished = "reconciliation finished, nothing to do"

	envBindWebValue  = "0.0.0.0:8025"
	envBindSmtpValue = "0.0.0.0:1025"

	envSmtpBind         = "MH_SMTP_BIND_ADDR"
	envApiBind          = "MH_API_BIND_ADDR"
	envUiBind           = "MH_UI_BIND_ADDR"
	envStorage          = "MH_STORAGE"
	envMongoUri         = "MH_MONGO_URI"
	envMongoDb          = "MH_MONGO_DB"
	envMongoCollection  = "MH_MONGO_COLLECTION"
	envMaildirPath      = "MH_MAILDIR_PATH"
	envHostname         = "MH_HOSTNAME"
	envCorsOrigin       = "MH_CORS_ORIGIN"
	envWebPath          = "MH_UI_WEB_PATH"
	envUpstreamSmtpFile = "MH_OUTGOING_SMTP"
	envWebAuthFile      = "MH_AUTH_FILE"

	argsJimInvite          = "-invite-jim"
	argsJimDisconnectRate  = "jim-disconnect"
	argsJimAccept          = "jim-accept"
	argsJimLinkSpeedAffect = "jim-linkspeed-affect"
	argsJimLinkSpeedMin    = "jim-linkspeed-min"
	argsJimLinkSpeedMax    = "jim-linkspeed-max"
	argsJimRejectSender    = "jim-reject-sender"
	argsJimRejectRecipient = "jim-reject-recipient"
	argsJimRejectAuth      = "jim-reject-auth"

	httpHealthPath = "/api/v2/messages?limit=1"
)
