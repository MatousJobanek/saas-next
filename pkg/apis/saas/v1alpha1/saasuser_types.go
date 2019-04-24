package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SaasUserSpec defines the desired state of SaasUser
type SaasUserSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	TargetClusterAddress string `json:"targetCluster,omitempty"`
	//Username             string `json:"username,omitempty"`
}

// SaasUserStatus defines the observed state of SaasUser
type SaasUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SaasUser is the Schema for the saasusers API
// +k8s:openapi-gen=true
type SaasUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SaasUserSpec   `json:"spec,omitempty"`
	Status SaasUserStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SaasUserList contains a list of SaasUser
type SaasUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SaasUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SaasUser{}, &SaasUserList{})
}
