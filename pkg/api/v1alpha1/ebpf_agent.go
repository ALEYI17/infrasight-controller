package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EbpfDaemonSetSpec struct {
	Image         string                      `json:"image"`
	NodeSelector  map[string]string           `json:"nodeSelector,omitempty"`
	Tolerations   []corev1.Toleration         `json:"tolerations,omitempty"`
	Resources     corev1.ResourceRequirements `json:"resources,omitempty"`
	RunPrivileged bool                        `json:"runPrivileged,omitempty"`
}

type EbpfDaemonSetStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type EbpfDaemonSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EbpfDaemonSetSpec   `json:"spec,omitempty"`
	Status EbpfDaemonSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type EbpfDaemonSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EbpfDaemonSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EbpfDaemonSet{}, &EbpfDaemonSetList{})
}
