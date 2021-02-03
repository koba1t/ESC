/*

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UserlandSpec defines the desired state of Userland
type UserlandSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of this resource. It used to naming owned resources.
	// +optional
	Name string `json:"Name,omitempty" protobuf:"bytes,1,opt,name=Name"`

	// TemplateName is the name of a Template in the same namespace as the binding this resource.
	TemplateName string `json:"templateName" protobuf:"bytes,2,opt,name=templateName"`

	// Enabled to create pod from userland resource.
	// Default true.
	// +optional
	Enabled bool `json:"enabled,omitempty" optional:"true" protobuf:"varint,3,opt,name=enabled"`
}

// UserlandStatus defines the observed state of Userland
type UserlandStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// Userland is the Schema for the userlands API
type Userland struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserlandSpec   `json:"spec,omitempty"`
	Status UserlandStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserlandList contains a list of Userland
type UserlandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Userland `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Userland{}, &UserlandList{})
}
