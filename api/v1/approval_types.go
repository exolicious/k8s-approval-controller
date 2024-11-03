package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

// ApprovalSpec defines the desired state of Approval
// +kubebuilder:validation:Required
type ApprovalSpec struct {
	// ResourceSpec holds the specification of the resource being wrapped.
	// This uses RawExtension to allow arbitrary resource specs to be embedded.
	// +kubebuilder:validation:Required
	// +kubebuilder:pruning:PreserveUnknownFields
	ResourceSpec runtime.RawExtension `json:"resourceSpec"`

	// Roles defines an array of roles that are allowed to approve this resource.
	// These roles should align with roles or groups in the identity provider (OIDC).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Roles []string `json:"roles"`

	// Metadata for additional information (e.g., human-readable names, descriptions).
	// +optional
	Meta map[string]string `json:"meta,omitempty"`
}

// ApprovalStatus defines the observed state of Approval
type ApprovalStatus struct {
	// Approved indicates whether the resource has been approved.
	// +kubebuilder:default="Pending"
	State string `json:"state"`
	// Active holds references to active instances created from this approval.
	// +optional
	ApprovedResource *corev1.ObjectReference `json:"active,omitempty"`
	DecisionTime     *metav1.Time            `json:"decisionTime,omitempty"`
	// Conditions provide detailed status information about the approval.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=apv
type Approval struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApprovalSpec   `json:"spec,omitempty"`
	Status ApprovalStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ApprovalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Approval `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Approval{}, &ApprovalList{})
}
