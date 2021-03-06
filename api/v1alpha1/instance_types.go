/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	StorageSetting       string
	TrafficInletResource string
)

const (
	// MemoryStorage incoming mails will be stored in process memory
	MemoryStorage StorageSetting = "memory"

	// MaildirStorage incoming mails will be stored in a folder
	MaildirStorage StorageSetting = "maildir"

	// MongoDBStorage incoming mails will be stored in a mongodb database
	MongoDBStorage StorageSetting = "mongodb"

	// RouteTrafficInlet an openshift route will be created to allow gui/api access
	RouteTrafficInlet TrafficInletResource = "route"

	// NoTrafficInlet no external access to the gui/api will be provided
	NoTrafficInlet TrafficInletResource = "none"

	// IngressTrafficInlet a k8s ingress will be created for gui/api access
	IngressTrafficInlet TrafficInletResource = "ingress"
)

// MailhogInstanceSpec defines the desired state of MailhogInstance
type MailhogInstanceSpec struct {
	// Image is the mailhog image to be used
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=4
	//+kubebuilder:default:="mailhog/mailhog:latest"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Mailhog Image",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Image string `json:"image,omitempty"`

	// Replicas is the count of pods to create
	//
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=10
	//+kubebuilder:validation:Required
	//+kubebuilder:default:=1
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Number of pods",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podCount"}
	Replicas int32 `json:"replicas,omitempty"`

	// Settings are mailhog configuration options, see https://github.com/mailhog/MailHog/blob/master/docs/CONFIG.md
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={storage:"memory"}
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Advanced Settings"
	Settings MailhogInstanceSettingsSpec `json:"settings,omitempty"`

	// WebTrafficInlet defines how the webinterface is exposed
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:default:="none"
	//+kubebuilder:validation:Enum=none;route;ingress
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Expose Mailhog with",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:route","urn:alm:descriptor:com.tectonic.ui:select:none","urn:alm:descriptor:com.tectonic.ui:select:ingress"}
	WebTrafficInlet TrafficInletResource `json:"webTrafficInlet,omitempty"`
}

// MailhogInstanceSettingsSpec are settings related to the mailhog instance
type MailhogInstanceSettingsSpec struct {
	// Hostname is the hostname for smtp ehlo/helo
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Format=hostname
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="SMTP Hostname",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Hostname string `json:"hostname,omitempty"`

	// CorsOrigin if set, this value is added into the Access-Control-Allow-Origin header returned by the API
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Web CORS AllowOrigin",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	CorsOrigin string `json:"corsOrigin,omitempty"`

	// Storage which storage backend to use, eg memory
	//
	//+kubebuilder:validation:Enum=memory;maildir;mongodb
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="memory"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Mail Storage Type",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:memory","urn:alm:descriptor:com.tectonic.ui:select:maildir","urn:alm:descriptor:com.tectonic.ui:select:mongodb"}
	Storage StorageSetting `json:"storage,omitempty"`

	// StorageMongoDb are only used when storage is set to mongodb
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="MongoDB Storage Settings",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.storage:mongodb"}
	StorageMongoDb MailhogStorageMongoDbSpec `json:"storageMongoDb,omitempty"`

	// StorageMaildir is only used when storage is set to maildir
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Maildir Storage Settings",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.storage:maildir"}
	StorageMaildir MailhogStorageMaildirSpec `json:"storageMaildir,omitempty"`

	// Files that configure more in-depth settings that require an additional configmap
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Mailhog Config Files"
	Files *MailhogFilesSpec `json:"files,omitempty"`

	// Resources allows to override the default resources of the created pods
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources reservations and limits",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Affinity allows to override the podtemplates affinity settings
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Affinity Settings"
	Affinity *AffinitySpec `json:"affinity,omitempty"`

	// Jim is the chaos monkey
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Jim / ChaosMonkey Config"
	Jim MailhogJimSpec `json:"jim,omitempty"`

	// WebPath context root under which web resources are served (without leading or trailing slashes), e.g. 'mailhog'
	// empty = no context root = serve all web resources under "/"
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Web ContextRoot",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	WebPath string `json:"webPath,omitempty"`

	// Ingress allows for k8s ingress related configuration
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Ingress Settings",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:fieldDependency:webTrafficInlet:ingress"}
	Ingress IngressSpec `json:"ingress,omitempty"`
}

// IngressSpec allows for k8s ingress related configuration
type IngressSpec struct {
	// Class will set the kubernetes.io/ingress.class of created k8s ingresses
	// leaving empty will use the default class
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Ingress Class",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:webTrafficInlet:ingress"}
	Class string `json:"class,omitempty"`

	// Host used for mailhog's ingress rule
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Hostname",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:webTrafficInlet:ingress"}
	Host string `json:"ingressClass,omitempty"`

	// TlsSecret which will be used for this ingress
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS Secret",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:webTrafficInlet:ingress"}
	TlsSecret string `json:"tlsSecret,omitempty"`
}

// AffinitySpec offers pod placement configuration
type AffinitySpec struct {
	// PodAffinity is used to get placed together with certain pods
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pod Affinity",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podAffinity"}
	PodAffinity *corev1.PodAffinity `json:"podAffinity,omitempty"`

	// PodAffinity is used to get placed apart from certain pods
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pod Anti Affinity",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:podAntiAffinity"}
	PodAntiAffinity *corev1.PodAntiAffinity `json:"podAntiAffinity,omitempty"`

	// NodeAffinity configures which nodes are preferred
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Node Affinity",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:nodeAffinity"}
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
}

// MailhogJimSpec invites jim into mailhog, the builtin chaos monkey
// they are added as args to the container's cmd
// see https://github.com/mailhog/MailHog/blob/master/docs/JIM.md
type MailhogJimSpec struct {
	// Invite set to true activates jim using the default values (see mh doc)
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Activate Chaosmonkey",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch"}
	Invite bool `json:"invite,omitempty"`

	// Disconnect Chance of randomly disconnecting a session (float, eg "0.005")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Connection Disconnect Chance",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	Disconnect string `json:"disconnect,omitempty"`

	// Accept Chance of accepting an incoming connection (float, eg "0.99")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Connection Accept Chance",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	Accept string `json:"accept,omitempty"`

	// LinkspeedAffect Chance of applying a rate limit (float, eg "0.1")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Connection Slow LinkSpeed Chance",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	LinkspeedAffect string `json:"linkspeedAffect,omitempty"`

	// LinkspeedMin Minimum link speed (in bytes per second, eg "1024")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Connection Slow LinkSpeed Minimum bytes/sec",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	LinkspeedMin string `json:"linkspeedMin,omitempty"`

	// LinkspeedMax Maximum link speed (in bytes per second, eg "10240")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Connection Slow LinkSpeed Maximum bytes/sec",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	LinkspeedMax string `json:"linkspeedMax,omitempty"`

	// RejectSender Chance of rejecting a MAIL FROM command (float, eg "0.05")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Chance the sender is rejected",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	RejectSender string `json:"rejectSender,omitempty"`

	// RejectRecipient Chance of rejecting a RCPT TO command (float, eg "0.05")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Chance the recipient is rejected",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	RejectRecipient string `json:"rejectRecipient,omitempty"`

	// RejectAuth Chance of rejecting an AUTH command (float, eg "0.05")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Chance the authentication is rejected",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:com.tectonic.ui:fieldDependency:settings.jim.invite:true"}
	RejectAuth string `json:"rejectAuth,omitempty"`
}

// MailhogFilesSpec is used to define settings that need to be passed as file (in a configmap)
type MailhogFilesSpec struct {
	// SmtpUpstreams Intercepted emails can be forwarded to upstreams via the UI
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="SMTP Upstreams for release"
	SmtpUpstreams []MailhogUpstreamSpec `json:"smtpUpstreams,omitempty"`

	// WebUsers If WebUsers are defined, UI/API Access will be protected with basic auth
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="HTTP Basic auth user restrictions"
	WebUsers []MailhogWebUserSpec `json:"webUsers,omitempty"`
}

// MailhogUpstreamSpec are upstream smtp servers a message can be release to that mailhog has intercepted (via gui/api)
// https://github.com/mailhog/MailHog-Server/blob/50f74a1aa2991b96313144d1ac718ce4d6739dfd/config/config.go#L55
type MailhogUpstreamSpec struct {
	// Name the Name this server will be shown under in the UI
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Server Name / Label",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Name string `json:"name,omitempty"`

	// Save is an option provided for compat reasons with mailhogs struct, just set it to true
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:default:=true
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Server Name / Label",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:hidden"}
	Save bool `json:"save,omitempty"`

	// Email the target Email address where the mail will be resent to
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=4
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Destination Email on release",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Email string `json:"email,omitempty"`

	// Host SMTP target Host hostname
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upstream SMTP server hostname",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Host string `json:"host,omitempty"`

	// Port SMTP target Port
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upstream SMTP server port",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Port string `json:"port,omitempty"`

	// Username the Username used for SMTP authentication
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upstream SMTP server username",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Username string `json:"username,omitempty"`

	// Password the Password used for SMTP authentication
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upstream SMTP server password",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Password string `json:"password,omitempty"`

	// Mechanism the SMTP login Mechanism used. This is _required_ when providing upstream user / password credentials
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength=4
	//+kubebuilder:validation:Enum=PLAIN;CRAMMD5
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upstream SMTP server auth mechanism",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:select:PLAIN","urn:alm:descriptor:com.tectonic.ui:select:CRAMMD5"}
	Mechanism string `json:"mechanism,omitempty"`
}

// MailhogWebUserSpec configures UI and API HTTP basic auth.
// see https://github.com/mailhog/MailHog/blob/master/docs/Auth.md for more information
type MailhogWebUserSpec struct {
	// Name is the username
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="HTTP Basic Auth Username",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Name string `json:"name,omitempty"`

	// PasswordHash is the bcrypt hash of the user's password
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=3
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Password bcrypt hash",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	PasswordHash string `json:"passwordHash,omitempty"`
}

// MailhogStorageMaildirSpec are settings applicable if the storage backend is maildir
// see https://github.com/mailhog/storage/blob/master/maildir.go for the implementation
type MailhogStorageMaildirSpec struct {
	// Path Maildir path (for maildir storage backend)
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=3
	//+kubebuilder:validation:Pattern:=`^(/)([\S]+(/)?)+$`
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Maildir path",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Path string `json:"path,omitempty"`

	// PvName if a PV name is given it will be used for maildir storage the pv needs to preexist, it will not be created
	// without a pv name an emptydir will be used which could lead to inconsistencies when multiple replicas are used
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=3
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="PV Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	PvName string `json:"pvName"`
}

// MailhogStorageMongoDbSpec are settings applicable if the storage backend is mongodb
// see https://github.com/mailhog/storage/blob/master/mongodb.go for the implementation
type MailhogStorageMongoDbSpec struct {
	// URI MongoDB host and port [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// for details about the URI format see https://pkg.go.dev/gopkg.in/mgo.v2#Dial
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=3
	//+kubebuilder:validation:Pattern:=`^(mongodb:(?:\/{2})?)((\w+?):(\w+?)@|:?@?)(\w+?):(\d+).*$`
	//+kubebuilder:validation:Format=uri
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="MongoDB URI",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	URI string `json:"uri,omitempty"`

	// Db MongoDB database name for message storage
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=2
	//+kubebuilder:validation:Pattern:=`^[\w-_]+$`
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="MongoDB DB",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Db string `json:"db,omitempty"`

	// Collection MongoDB collection name for message storage
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=2
	//+kubebuilder:validation:Pattern:=`^[\w-_]+$`
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="MongoDB Collection",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Collection string `json:"collection,omitempty"`
}

// MailhogInstanceStatus defines the observed state of MailhogInstance
type MailhogInstanceStatus struct {
	// Pods all the podnames owned by the cr
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pod Status",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Pods PodStatus `json:"pods,omitempty"`

	// PodCount is the amount of last seen pods belonging to this cr
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pod Count",xDescriptors="urn:alm:descriptor:com.tectonic.ui:podCount"
	PodCount int `json:"podCount,omitempty"`

	// ReadyPodCount is the amount of pods last seen ready
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Ready Pod Count"
	ReadyPodCount int `json:"readyPodCount,omitempty"`

	// LabelSelector is the labelselector which can be used by HPA
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Label Selector"
	LabelSelector string `json:"labelSelector,omitempty"`

	// Error is used to signal illegal CR specs
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Error Message"
	Error string `json:"error,omitempty"`

	// RouteURL will be set to the path under which mailhog is reachable if openshift Route is enabled
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Mailhog Web UI",xDescriptors="urn:alm:descriptor:org.w3:link"
	RouteURL string `json:"routeUrl,omitempty"`
}

// PodStatus will divide the child pods into a grouping
type PodStatus struct {
	// Pending contains pod names where the podStatus is pending
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Pending []string `json:"pending,omitempty"`

	// Failed contains pod names where the podStatus is failed
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Failed []string `json:"failed,omitempty"`

	// Ready contains pod names where the mailhog container is in ready state
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Ready []string `json:"ready,omitempty"`

	// Restarting contains pod names where the mailhog container restarted more than 3 times
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Restarting []string `json:"restarting,omitempty"`

	// Other contains pod names where the condition is not caught by the other states (e.g. pod running bot container not ready)
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Other []string `json:"other,omitempty"`
}

// MailhogInstance is the Schema for the mailhoginstances API
//
//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.podCount,selectorpath=.status.labelSelector
//+operator-sdk:csv:customresourcedefinitions:displayName="Mailhog Instance"
//+operator-sdk:csv:customresourcedefinitions:resources={{Service,v1},{Deployment,v1},{Route,v1},{ConfigMap,v1},{Ingress,v1}}
type MailhogInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MailhogInstanceSpec `json:"spec,omitempty"`

	// Status last observed status
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Status MailhogInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MailhogInstanceList contains a list of MailhogInstance
type MailhogInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MailhogInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MailhogInstance{}, &MailhogInstanceList{})
}
