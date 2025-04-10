/*
Copyright 2025.

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

// EbpfDaemonSetSpec defines the desired state of EbpfDaemonSet.
type EbpfDaemonSetSpec struct {
	Image         string                      `json:"image"`
	NodeSelector  map[string]string           `json:"nodeSelector,omitempty"`
	Tolerations   []corev1.Toleration         `json:"tolerations,omitempty"`
	Resources     corev1.ResourceRequirements `json:"resources,omitempty"`
	RunPrivileged bool                        `json:"runPrivileged,omitempty"`
}

// EbpfDaemonSetStatus defines the observed state of EbpfDaemonSet.
type EbpfDaemonSetStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="ConditionType",type=string,JSONPath=`.status.conditions[0].type`
// +kubebuilder:printcolumn:name="ConditionStatus",type=string,JSONPath=`.status.conditions[0].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[0].reason`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.conditions[0].message`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// EbpfDaemonSet is the Schema for the ebpfdaemonsets API.
type EbpfDaemonSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EbpfDaemonSetSpec   `json:"spec,omitempty"`
	Status EbpfDaemonSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EbpfDaemonSetList contains a list of EbpfDaemonSet.
type EbpfDaemonSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EbpfDaemonSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EbpfDaemonSet{}, &EbpfDaemonSetList{})
}
