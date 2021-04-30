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
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var namespaceclaimlog = logf.Log.WithName("namespaceclaim-resource")

func (r *NamespaceClaim) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-claim-tmax-io-v1alpha1-namespaceclaim,mutating=false,failurePolicy=fail,groups=claim.tmax.io,resources=namespaceclaims;namespaceclaims/status,versions=v1alpha1,name=vnamespaceclaim.kb.io

var _ webhook.Validator = &NamespaceClaim{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceClaim) ValidateCreate() error {
	namespaceclaimlog.Info("validate create", "name", r.Name)
	namespaceclaimlog.Info("validating Webhook for NamespaceClaim CRD Start!!")
	return r.validateNscRq()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceClaim) ValidateUpdate(old runtime.Object) error {
	namespaceclaimlog.Info("validate update", "name", r.Name)
	// TODO(user): fill in your validation logic upon object update.
	old_status := old.(*NamespaceClaim).DeepCopy().Status.Status
	now_status := r.Status.Status

	if (old_status == NamespaceClaimStatusTypeSuccess && (now_status != NamespaceClaimStatusTypeSuccess && now_status != NamespaceClaimStatueTypeDeleted && now_status != NamespaceClaimStatueTypeError)) ||
		(old_status == NamespaceClaimStatueTypeDeleted && now_status != NamespaceClaimStatueTypeDeleted) {
		return errors.NewForbidden(
			schema.GroupResource{Group: "claim.tmax.io", Resource: r.Name},
			"",
			err.New("cannot update NamespaceClaim in Success or Deleted status"),
		)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NamespaceClaim) ValidateDelete() error {
	namespaceclaimlog.Info("validate delete", "name", r.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *NamespaceClaim) validateNscRq() error {
	var allErrs field.ErrorList

	if err := r.validateNscRqSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return errors.NewInvalid(schema.GroupKind{Group: "claim.tmax.io", Kind: "namespaceclaim"}, "resourceQuotaSpecName", allErrs)
}

func (r *NamespaceClaim) validateNscRqSpec() *field.Error {
	checkRequireNameList := []string{}
	for resourceName, _ := range r.Spec.Hard {
		if !contains(ResourceNameList, resourceName.String()) {
			return field.Invalid(field.NewPath(resourceName.String()), resourceName.String(), "Invalid ResourceQuotaSpecName")
		}
		checkRequireNameList = append(checkRequireNameList, resourceName.String())
	}
	// if !(contains(checkRequireNameList, string(v1.ResourceLimitsCPU)) && contains(checkRequireNameList, string(v1.ResourceLimitsMemory))) {
	// 	return field.Invalid(nil, nil, "limits.cpu & limits.memory are Mandatory")
	// }
	return nil
}
