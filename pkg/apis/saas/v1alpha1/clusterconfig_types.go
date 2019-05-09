package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterConfigSpec defines the desired state of ClusterConfig
type ClusterConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Config SaasClusterRoleConfig `json:"config,omitempty"`
}

type SaasClusterRoleConfig struct {
	ApiAddress string              `json:"apiAddress"`
	Role       SaasClusterRole     `json:"role"`
	Host       SaasClusterConfig   `json:"host,omitempty"`
	Members    []SaasClusterConfig `json:"members,omitempty"`
}

type SaasClusterRole string

const Host SaasClusterRole = "host"
const Member SaasClusterRole = "member"

type SaasClusterConfig struct {
	ApiAddress string `json:"apiAddress"`
	// SecretRef refers to the secret that contains credentials to access the git repo. Optional.
	SecretRef *SecretRef `json:"secretRef,omitempty"`
}

// SecretRef holds information about the secret that contains credentials to access the git repo
type SecretRef struct {
	// Name is the name of the secret that contains credentials to access the git repo
	Name string `json:"name"`
}

// ClusterConfigStatus defines the observed state of ClusterConfig
type ClusterConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterConfig is the Schema for the clusterconfigs API
// +k8s:openapi-gen=true
type ClusterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterConfigSpec   `json:"spec,omitempty"`
	Status ClusterConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterConfigList contains a list of ClusterConfig
type ClusterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterConfig{}, &ClusterConfigList{})
}
