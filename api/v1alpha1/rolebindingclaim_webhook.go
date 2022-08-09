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
	err "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var rolebindingclaimlog = logf.Log.WithName("rolebindingclaim-resource")

func (r *RoleBindingClaim) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:verbs=create;update,path=/validate-claim-tmax-io-v1alpha1-rolebindingclaim,mutating=false,failurePolicy=fail,groups=claim.tmax.io,resources=rolebindingclaims/status,versions=v1alpha1,name=vrolebindingclaim.kb.io,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun
//+kubebuilder:webhook:verbs=create;update,path=/validate-claim-tmax-io-v1alpha1-namespaceclaim,mutating=false,failurePolicy=fail,groups=claim.tmax.io,resources=namespaceclaims;namespaceclaims/status,versions=v1alpha1,name=vnamespaceclaim.kb.io,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun

var _ webhook.Validator = &RoleBindingClaim{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *RoleBindingClaim) ValidateCreate() error {
	rolebindingclaimlog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *RoleBindingClaim) ValidateUpdate(old runtime.Object) error {
	rolebindingclaimlog.Info("validate update", "name", r.Name)

	old_status := old.(*RoleBindingClaim).DeepCopy().Status.Status
	now_status := r.Status.Status

	if (old_status == RoleBindingClaimStatusTypeSuccess && (now_status != RoleBindingClaimStatusTypeSuccess && now_status != RoleBindingClaimStatusTypeDeleted && now_status != RoleBindingClaimStatusTypeError)) ||
		(old_status == RoleBindingClaimStatusTypeDeleted && now_status != RoleBindingClaimStatusTypeDeleted) {
		return errors.NewForbidden(
			schema.GroupResource{Group: "claim.tmax.io", Resource: r.Name},
			"",
			err.New("Cannot update RoleBindingClaim in Approved or Deleted status"),
		)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *RoleBindingClaim) ValidateDelete() error {
	rolebindingclaimlog.Info("validate delete", "name", r.Name)
	if r.Status.Status == RoleBindingClaimStatusTypeSuccess {
		return err.New("Cannot delete RoleBindingClaim before deleting Rolebinding")
	}
	return nil
}
