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

// Important: Run "make" to regenerate code after modifying this file

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	StorageSetting       string
	BackingResource      string
	TrafficInletResource string
)

const (
	MemoryStorage  StorageSetting = "memory"
	MaildirStorage StorageSetting = "maildir"
	MongoDBStorage StorageSetting = "mongodb"

	DeploymentBacking       BackingResource = "deployment"
	DeploymentConfigBacking BackingResource = "deploymentConfig"

	RouteTrafficInlet TrafficInletResource = "route"
	NoTrafficInlet    TrafficInletResource = "none"
)

// MailhogInstanceSpec defines the desired state of MailhogInstance
type MailhogInstanceSpec struct {
	// Image is the mailhog image to be used
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=4
	//+kubebuilder:default:="mailhog/mailhog:latest"
	Image string `json:"image,omitempty"`

	// Replicas is the count of pods to create
	//
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=10
	//+kubebuilder:validation:Required
	//+kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`

	// Settings are mailhog configuration options, see https://github.com/mailhog/MailHog/blob/master/docs/CONFIG.md
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={storage:"memory"}
	Settings MailhogInstanceSettingsSpec `json:"settings,omitempty"`

	// WebTrafficInlet defines how the webinterface is exposed
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:default:="none"
	//+kubebuilder:validation:Enum=none;route
	WebTrafficInlet TrafficInletResource `json:"webTrafficInlet,omitempty"`

	// BackingResource controls if a deploymentConfig or deployment is used
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:default:="deployment"
	//+kubebuilder:validation:Enum=deployment;deploymentConfig
	BackingResource BackingResource `json:"backingResource,omitempty"`
}

// MailhogInstanceSettingsSpec are settings related to the mailhog instance
type MailhogInstanceSettingsSpec struct {
	// Hostname is the hostname for smtp ehlo/helo
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Format=hostname
	//+optional
	//+nullable
	Hostname string `json:"hostname,omitempty"`

	// CorsOrigin if set, this value is added into the Access-Control-Allow-Origin header returned by the API
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	CorsOrigin string `json:"corsOrigin,omitempty"`

	// Storage which storage backend to use, eg memory
	//
	//+kubebuilder:validation:Enum=memory;maildir;mongodb
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="memory"
	Storage StorageSetting `json:"storage,omitempty"`

	// StorageMongoDb are only used when storage is set to mongodb
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	StorageMongoDb MailhogStorageMongoDbSpec `json:"storageMongoDb,omitempty"`

	// StorageMaildir is only used when storage is set to maildir
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	StorageMaildir MailhogStorageMaildirSpec `json:"storageMaildir,omitempty"`

	// Files that configure more in-depth settings that require an additional configmap
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Files *MailhogFilesSpec `json:"files,omitempty"`

	// Jim is the chaos monkey
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Jim MailhogJimSpec `json:"jim,omitempty"`
}

// MailhogJimSpec invites jim into mailhog, the builtin chaos monkey
// see https://github.com/mailhog/MailHog/blob/master/docs/JIM.md
type MailhogJimSpec struct {
	// Invite set to true activates jim using the default values (see mh doc)
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	//+optional
	//+nullable
	Invite bool `json:"invite,omitempty"`

	// Disconnect Chance of randomly disconnecting a session (float, eg "0.005")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Disconnect string `json:"disconnect,omitempty"`

	// Accept Chance of accepting an incoming connection (float, eg "0.99")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Accept string `json:"accept,omitempty"`

	// LinkspeedAffect Chance of applying a rate limit (float, eg "0.1")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	LinkspeedAffect string `json:"linkspeedAffect,omitempty"`

	// LinkspeedMin Minimum link speed (in bytes per second, eg "1024")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	LinkspeedMin string `json:"linkspeedMin,omitempty"`

	// LinkspeedMax Maximum link speed (in bytes per second, eg "10240")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	LinkspeedMax string `json:"linkspeedMax,omitempty"`

	// RejectSender Chance of rejecting a MAIL FROM command (float, eg "0.05")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	RejectSender string `json:"rejectSender,omitempty"`

	// RejectRecipient Chance of rejecting a RCPT TO command (float, eg "0.05")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	RejectRecipient string `json:"rejectRecipient,omitempty"`

	// RejectAuth Chance of rejecting an AUTH command (float, eg "0.05")
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	RejectAuth string `json:"rejectAuth,omitempty"`
}

type MailhogFilesSpec struct {
	// SmtpUpstreams Intercepted emails can be forwarded to upstreams via the UI
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	SmtpUpstreams []MailhogUpstreamSpec `json:"smtpUpstreams,omitempty"`

	// WebUsers If WebUsers are defined, UI/API Access will be protected with basic auth
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	WebUsers []MailhogWebUserSpec `json:"webUsers,omitempty"`
}

type MailhogUpstreamSpec struct {
	// Name the Name this server will be shown under in the UI
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	Name string `json:"name,omitempty"`

	// Save is an option provided for compat reasons with mailhogs struct, just set it to true
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:default:=true
	Save bool `json:"save,omitempty"`

	// Email the target Email address where the mail will be resent to
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=4
	Email string `json:"email,omitempty"`

	// Host SMTP target Host hostname
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	Host string `json:"host,omitempty"`

	// Port SMTP target Port
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	Port string `json:"port,omitempty"`

	// Username the Username used for SMTP authentication
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Username string `json:"username,omitempty"`

	// Password the Password used for SMTP authentication
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Password string `json:"password,omitempty"`

	// Mechanism the SMTP login Mechanism used. This is _required_ when providing upstream user / password credentials
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength=4
	//+kubebuilder:validation:Enum=PLAIN;CRAMMD5
	//+optional
	//+nullable
	Mechanism string `json:"mechanism,omitempty"`
}

type MailhogWebUserSpec struct {
	// Name is the username
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=2
	Name string `json:"name,omitempty"`

	// PasswordHash is the bcrypt hash of the user's password
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=3
	PasswordHash string `json:"passwordHash,omitempty"`
}

// MailhogStorageMaildirSpec are settings applicable if the storage backend is maildir
type MailhogStorageMaildirSpec struct {
	// Path Maildir path (for maildir storage backend)
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=3
	//+kubebuilder:validation:Pattern:=`^(/)([\S]+(/)?)+$`
	//+optional
	//+nullable
	Path string `json:"path,omitempty"`
}

// MailhogStorageMongoDbSpec are settings applicable if the storage backend is mongodb
type MailhogStorageMongoDbSpec struct {
	// URI MongoDB host and port
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=3
	//+kubebuilder:validation:Pattern:=`^(mongodb:(?:\/{2})?)((\w+?):(\w+?)@|:?@?)(\w+?):(\d+).*$`
	//+kubebuilder:validation:Format=uri
	//+optional
	//+nullable
	URI string `json:"uri,omitempty"`

	// Db MongoDB database name for message storage
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=2
	//+kubebuilder:validation:Pattern:=`^[\w-_]+$`
	//+optional
	//+nullable
	Db string `json:"db,omitempty"`

	// Collection MongoDB collection name for message storage
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MinLength:=2
	//+kubebuilder:validation:Pattern:=`^[\w-_]+$`
	//+optional
	//+nullable
	Collection string `json:"collection,omitempty"`
}

// MailhogInstanceStatus defines the observed state of MailhogInstance
type MailhogInstanceStatus struct {
	// Pods all the podnames owned by the cr
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	Pods []string `json:"pods,omitempty"`

	// PodCount is the amount of last seen pods belonging to this cr
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	PodCount int `json:"podCount,omitempty"`

	// LabelSelector is the labelselector which can be used by HPA
	//
	//+kubebuilder:validation:Optional
	//+optional
	//+nullable
	LabelSelector string `json:"labelSelector,omitempty"`
}

// MailhogInstance is the Schema for the mailhoginstances API
//
//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.podCount,selectorpath=.status.labelSelector
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
