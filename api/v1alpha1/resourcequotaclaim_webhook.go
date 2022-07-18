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

	"github.com/tmax-cloud/hypercloud-single-operator/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var ResourceNameList = []string{
	string(v1.ResourceCPU),
	string(v1.ResourceMemory),
	string(v1.ResourceEphemeralStorage),
	string(v1.ResourceRequestsCPU),
	string(v1.ResourceRequestsMemory),
	string(v1.ResourceRequestsEphemeralStorage),
	string(v1.ResourceLimitsCPU),
	string(v1.ResourceLimitsMemory),
	string(v1.ResourceLimitsEphemeralStorage),
	string(v1.ResourcePods),
	string(v1.ResourceQuotas),
	string(v1.ResourceServices),
	string(v1.ResourceReplicationControllers),
	string(v1.ResourceSecrets),
	string(v1.ResourceConfigMaps),
	string(v1.ResourcePersistentVolumeClaims),
	string(v1.ResourceStorage),
	string(v1.ResourceRequestsStorage),
	string(v1.ResourceServicesNodePorts),
	string(v1.ResourceServicesLoadBalancers),
	"count/" + string(v1.ResourcePersistentVolumeClaims),
	"count/" + string(v1.ResourceServices),
	"count/" + string(v1.ResourceConfigMaps),
	"count/" + string(v1.ResourceReplicationControllers),
	"count/deployments.apps",
	"count/replicasets.apps",
	"count/statefulsets.apps",
	"count/jobs.batch",
	"count/cronjobs.batch",
	"count/deployments.extensions",
	"requests.nvidia.com/gpu",
	"ssd-ceph-fs.storageclass.storage.k8s.io/requests.storage",
	"hdd-ceph-fs.storageclass.storage.k8s.io/requests.storage",
	"ssd-ceph-block.storageclass.storage.k8s.io/requests.storage",
	"hdd-ceph-block.storageclass.storage.k8s.io/requests.storage",
}

// log is for logging in this package.
var resourcequotaclaimlog = logf.Log.WithName("resourcequotaclaim-resource")

func (r *ResourceQuotaClaim) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-claim-tmax-io-v1alpha1-resourcequotaclaim,mutating=false,failurePolicy=fail,groups=claim.tmax.io,resources=resourcequotaclaims;resourcequotaclaims/status,versions=v1alpha1,name=vresourcequotaclaim.kb.io,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun
// +kubebuilder:webhook:verbs=create;update,path=/validate-claim-tmax-io-v1alpha1-namespaceclaim,mutating=false,failurePolicy=fail,groups=claim.tmax.io,resources=namespaceclaims;namespaceclaims/status,versions=v1alpha1,name=vnamespaceclaim.kb.io,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun,admissionReviewVersions=v1beta1;v1,sideEffects=NoneOnDryRun

var _ webhook.Validator = &ResourceQuotaClaim{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ResourceQuotaClaim) ValidateCreate() error {
	resourcequotaclaimlog.V(3).Info("validate create", "name", r.Name)
	resourcequotaclaimlog.V(3).Info("validating Webhook for resourcequotaClaim CRD Start!!")
	return r.validateRqc()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ResourceQuotaClaim) ValidateUpdate(old runtime.Object) error {
	resourcequotaclaimlog.V(3).Info("validate update", "name", r.Name)
	// TODO(user): fill in your validation logic upon object update.
	old_status := old.(*ResourceQuotaClaim).DeepCopy().Status.Status
	now_status := r.Status.Status

	if (old_status == ResourceQuotaClaimStatusTypeSuccess && (now_status != ResourceQuotaClaimStatusTypeSuccess && now_status != ResourceQuotaClaimStatusTypeDeleted && now_status != ResourceQuotaClaimStatusTypeError)) ||
		(old_status == ResourceQuotaClaimStatusTypeDeleted && now_status != ResourceQuotaClaimStatusTypeDeleted) {
		return errors.NewForbidden(
			schema.GroupResource{Group: "claim.tmax.io", Resource: r.Name},
			"",
			err.New("cannot update ResourceQuotaClaim in Approved or Deleted status"),
		)
	}

	return r.validateRqc()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ResourceQuotaClaim) ValidateDelete() error {
	resourcequotaclaimlog.V(3).Info("validate delete", "name", r.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *ResourceQuotaClaim) validateRqc() error {
	var allErrs field.ErrorList

	if err := r.validateRqcSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return errors.NewInvalid(schema.GroupKind{Group: "claim.tmax.io", Kind: "resourcequotaclaim"}, "resourceQuotaSpecName", allErrs)
}

func (r *ResourceQuotaClaim) validateRqcSpec() *field.Error {
	checkRequireNameList := []string{}
	for resourceName := range r.Spec.Hard {
		if !contains(ResourceNameList, resourceName.String()) {
			if !contains(util.ResourceList, resourceName.String()) {
				util.UpdateResourceList(resourcequotaclaimlog)
				if !contains(util.ResourceList, resourceName.String()) {
					return field.Invalid(field.NewPath(resourceName.String()), resourceName.String(), "Invalid ResourceQuotaSpecName")
				}
			}
		}
		checkRequireNameList = append(checkRequireNameList, resourceName.String())
	}
	if !(contains(checkRequireNameList, string(v1.ResourceLimitsCPU)) && contains(checkRequireNameList, string(v1.ResourceLimitsMemory))) {
		return field.Invalid(nil, nil, "limits.cpu & limits.memory are Mandatory")
	}
	return nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
