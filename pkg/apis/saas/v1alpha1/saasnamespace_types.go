package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SaasNamespaceSpec defines the desired state of SaasNamespace
type SaasNamespaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Owner         string
	NamespaceName string
}

// SaasNamespaceStatus defines the observed state of SaasNamespace
type SaasNamespaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SaasNamespace is the Schema for the saasnamespaces API
// +k8s:openapi-gen=true
type SaasNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SaasNamespaceSpec   `json:"spec,omitempty"`
	Status SaasNamespaceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SaasNamespaceList contains a list of SaasNamespace
type SaasNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SaasNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SaasNamespace{}, &SaasNamespaceList{})
}
