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

package v1alpha1

import (
	rbacApi "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RolebindingClaimStatusType defines Rolebindingclaim status type
// type RoleBindingClaimStatusType string

const (
	RoleBindingClaimStatusTypeAwaiting = "Awaiting"
	RoleBindingClaimStatusTypeSuccess  = "Approved"
	RoleBindingClaimStatusTypeReject   = "Rejected"
	RoleBindingClaimStatusTypeError    = "Error"
	RoleBindingClaimStatusTypeDeleted  = "Role Binding Deleted"
)

// RoleBindingClaimStatus defines the observed state of RoleBindingClaim
type RoleBindingClaimStatus struct {
	// Message shows log when the status changed in last
	Message string `json:"message,omitempty" protobuf:"bytes,1,opt,name=message"`
	// Reason shows why the status changed in last
	Reason string `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`
	// LastTransitionTime shows the time when the status changed in last
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// +kubebuilder:validation:Enum=Awaiting;Approved;Rejected;Error;Role Binding Deleted;
	// Status shows the present status of the NamespaceClaim
	Status string `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rbc
// +kubebuilder:printcolumn:name="ResourceName",type=string,JSONPath=`.resourceName`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.reason`
// RoleBindingClaim is the Schema for the rolebindingclaims API
type RoleBindingClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=.metadata"`
	// rolebindingName is name of the role binding which will be created
	ResourceName string `json:"resourceName"`
	// Subjects contains a reference to the object or user identities a role binding applies to.  This can either hold a direct API object reference,
	// or a value for non-objects such as user and group names.
	Subjects []rbacApi.Subject `json:"subjects,omitempty" protobuf:"bytes,2,rep,name=subjects"`
	// RoleRef contains information that points to the role being used
	RoleRef rbacApi.RoleRef `json:"roleRef" protobuf:"bytes,3,opt,name=roleRef"`
	// Status shows the present status of the RoleBindingClaim
	Status RoleBindingClaimStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoleBindingClaimList contains a list of RoleBindingClaim
type RoleBindingClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBindingClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoleBindingClaim{}, &RoleBindingClaimList{})
}
