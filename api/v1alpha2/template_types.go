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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//VolumeSpec defines the volume of TemplateSpec
type VolumeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//VolumeName is unified volume name.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	VolumeMount VolumeMount `json:"volumeMount" protobuf:"bytes,2,opt,name=volumeMount"`

	PVCSpec v1.PersistentVolumeClaimSpec `json:"pvcSpec" protobuf:"bytes,3,opt,name=pvcSpec"`
}

//VolumeMount defines the mount point for volume. (like v1.VolumeMount)
type VolumeMount struct {
	//ContainerName match VolumeSpec to container with name at template.spec.conatner.name field.
	ContainerName string `json:"containerName" protobuf:"bytes,1,opt,name=containerName"`

	MountPath string `json:"mountPath" protobuf:"bytes,3,opt,name=mountPath"`

	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`
}

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Template v1.PodTemplateSpec `json:"template" protobuf:"bytes,1,opt,name=template"`

	ServiceSpec v1.ServiceSpec `json:"service,omitempty" protobuf:"bytes,2,opt,name=service"`

	VolumeSpecs []VolumeSpec `json:"volumes,omitempty" protobuf:"bytes,3,opt,name=volumes"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}