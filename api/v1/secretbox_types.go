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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretboxSpec defines the desired state of Secretbox
type SecretboxSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=1
	Size int32 `json:"size"`
}

// SecretboxStatus defines the observed state of Secretbox
type SecretboxStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	State SecretboxState `json:"state,omitempty"`
}

type SecretboxState string

const (
	SecretboxStateHealthy   SecretboxState = "Healthy"
	SecretboxStateUnhealthy SecretboxState = "Unhealthy"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Secretbox is the Schema for the secretboxes API
type Secretbox struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretboxSpec   `json:"spec,omitempty"`
	Status SecretboxStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretboxList contains a list of Secretbox
type SecretboxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secretbox `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Secretbox{}, &SecretboxList{})
}

// GetSecretName returns the name of the secret
func (s *Secretbox) GetSecretName() string {
	return "secretbox-" + s.Name + "-secret"
}

// GetConfigMapName returns the name of the configmap
func (s *Secretbox) GetConfigMapName() string {
	return "secretbox-" + s.Name + "-configmap"
}

// GetLabels returns the labels of the secretbox
func (s *Secretbox) GetLabels() map[string]string {
	return map[string]string{
		"app": "secretbox-" + s.Name,
	}
}

// GetDeploymentName returns the name of the deployment
func (s *Secretbox) GetDeploymentName() string {
	return "secretbox-" + s.Name + "-deployment"
}

// GetServiceName returns the name of the service
func (s *Secretbox) GetServiceName() string {
	return "secretbox-" + s.Name + "-service"
}
