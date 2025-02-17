/*
Copyright 2021.

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
	"github.com/operator-framework/operator-lib/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConditionReconciled is a status condition type that indicates whether the
	// CR has been successfully reconciled
	ConditionReconciled status.ConditionType = "Reconciled"
	// ReconciledReasonComplete indicates the CR was successfully reconciled
	ReconciledReasonComplete status.ConditionReason = "ReconcileComplete"
	// ReconciledReasonError indicates an error was encountered while
	// reconciling the CR
	ReconciledReasonError status.ConditionReason = "ReconcileError"
)

type ExtractSpec struct {
	Branch string `json:"branch"`
	Repo   string `json:"repo"`
	Email  string `json:"email"`
	Secret string `json:"secret"`
}

// ExtractStatus defines the observed state of Extract
type ExtractStatus struct {
	Completed  bool              `json:"completed,omitempty"`
	Conditions status.Conditions `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Extract is the Schema for the extracts API
type Extract struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtractSpec   `json:"spec,omitempty"`
	Status ExtractStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExtractList contains a list of Extract
type ExtractList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Extract `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Extract{}, &ExtractList{})
}
