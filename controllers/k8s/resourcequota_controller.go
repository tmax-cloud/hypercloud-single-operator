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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ResourceQuotaReconciler reconciles a ResourceQuota object
type ResourceQuotaReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *ResourceQuotaReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log

	reqLogger.Info("Reconciling ResourceQuota")

	r.replaceRQCStatus(req.Name, req.Namespace, claim.ResourceQuotaClaimStatusTypeDeleted)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceQuotaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ResourceQuota{}).
		WithEventFilter(
			predicate.Funcs{
				// Reconciling only when ResourceQuota is deleted
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					return false
				},
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
			},
		).
		Complete(r)
}

func (r *ResourceQuotaReconciler) replaceRQCStatus(rqcName string, rqNamespace string, status string) {
	reqLogger := r.Log
	rqcFound := &claim.ResourceQuotaClaim{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: rqcName, Namespace: rqNamespace}, rqcFound); err != nil && errors.IsNotFound(err) {
		reqLogger.Info("ResourceQuotaClaim [ " + rqcName + " ] Not Exists, Do Nothing")
	} else {
		rqcFound.Status.Status = status
		rqcFound.Status.Reason = "ResourceQuota [ " + rqcName + " ] Deleted"
		if err := r.Status().Update(context.TODO(), rqcFound); err != nil {
			reqLogger.Error(err, "Failed to Update ResourceQuotaClaim [ "+rqcName+" ]")
			panic("Failed to Update ResourceQuotaClaim [ " + rqcName + " ]")
		} else {
			reqLogger.Info("Update ResourceQuotaClaim [ " + rqcName + " ] Success")
		}
	}
}