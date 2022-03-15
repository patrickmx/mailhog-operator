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
	WebTrafficInlet string `json:"webTrafficInlet,omitempty"`
}

// MailhogInstanceSettingsSpec are settings related to the mailhog instance
type MailhogInstanceSettingsSpec struct {

	// Hostname is the hostname for smtp ehlo/helo
	//
	//+kubebuilder:validation:Optional
	Hostname string `json:"hostname,omitempty"`

	// CorsOrigin if set, this value is added into the Access-Control-Allow-Origin header returned by the API
	//
	//+kubebuilder:validation:Optional
	CorsOrigin string `json:"corsOrigin,omitempty"`

	// Storage which storage backend to use, eg memory
	//
	//+kubebuilder:validation:Enum=memory;maildir;mongodb
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="memory"
	Storage string `json:"storage,omitempty"`

	// StorageMongoDb are only used when storage is set to mongodb
	//
	//+kubebuilder:validation:Optional
	StorageMongoDb MailhogStorageMongoDbSpec `json:"storageMongoDb,omitempty"`

	// StorageMaildir is only used when storage is set to maildir
	//
	//+kubebuilder:validation:Optional
	StorageMaildir MailhogStorageMaildirSpec `json:"storageMaildir,omitempty"`
}

// MailhogStorageMaildirSpec are settings applicable if the storage backend is maildir
type MailhogStorageMaildirSpec struct {

	// Path Maildir path (for maildir storage backend)
	//
	//+kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`
}

// MailhogStorageMongoDbSpec are settings applicable if the storage backend is mongodb
type MailhogStorageMongoDbSpec struct {

	// Uri MongoDB host and port
	//
	//+kubebuilder:validation:Optional
	Uri string `json:"uri,omitempty"`

	// Db MongoDB database name for message storage
	//
	//+kubebuilder:validation:Optional
	Db string `json:"db,omitempty"`

	// Collection MongoDB collection name for message storage
	//
	//+kubebuilder:validation:Optional
	Collection string `json:"collection,omitempty"`
}

// MailhogInstanceStatus defines the observed state of MailhogInstance
type MailhogInstanceStatus struct {

	// Pods all the podnames owned by the cr
	//
	Pods []string `json:"pods,omitempty"`

	// PodCount is the amount of last seen pods belonging to this cr
	//
	PodCount int `json:"podCount,omitempty"`

	// LabelSelector is the labelselector which can be used by HPA
	//
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

	Spec   MailhogInstanceSpec   `json:"spec,omitempty"`
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
