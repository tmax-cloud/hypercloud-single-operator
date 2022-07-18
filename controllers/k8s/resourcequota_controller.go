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
	"fmt"

	"github.com/go-logr/logr"
	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
	"github.com/tmax-cloud/hypercloud-single-operator/util"
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

func (r *ResourceQuotaReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log

	reqLogger.V(3).Info("Reconciling ResourceQuota")
	resourcequota := &v1.ResourceQuota{}

	if err := r.Get(context.TODO(), req.NamespacedName, resourcequota); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.V(3).Info("ResourceQuota resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		} else {
			reqLogger.V(1).Error(err, "Failed to get ResourceQuota")
			return ctrl.Result{}, err
		}
	}

	//set helper
	if helper, err := patch.NewHelper(resourcequota, r.Client); err != nil {
		return ctrl.Result{}, err
	} else {
		r.patchHelper = helper
	}
	defer func() {
		r.patchHelper.Patch(context.TODO(), resourcequota)
	}()

	defer func() {
		s := recover()
		if s != nil {
			fmt.Println("Error !! : ", s)
		}
	}()

	if resourcequota.DeletionTimestamp != nil {
		if resourcequota.Labels != nil && resourcequota.Labels["fromClaim"] != "" {
			if resourcequota.Finalizers != nil {
				resourcequota.Finalizers = util.RemoveValue(resourcequota.Finalizers, "resourcequota/finalizers")
				reqLogger.V(3).Info("Delete Finalizer [ resourcequota/finalizers ] Success")
			}

			r.replaceRQCStatus(resourcequota.Labels["fromClaim"], resourcequota.Name, resourcequota.Namespace, claim.ResourceQuotaClaimStatusTypeDeleted)
			reqLogger.V(3).Info("Update ResourceQuotaClaim [ " + resourcequota.Labels["fromClaim"] + " ] Status to ResourceQuota Deleted")
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceQuotaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ResourceQuota{}).
		WithEventFilter(
			predicate.Funcs{
				// Reconciling only when ResourceQuota is deleted,
				// which is noticed by DeletionTimestamp
				DeleteFunc: func(e event.DeleteEvent) bool {
					return false
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					newRQ := e.ObjectNew.(*v1.ResourceQuota).DeepCopy()
					if newRQ.DeletionTimestamp != nil {
						return true
					}
					return false
				},
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
			},
		).
		Complete(r)
}

func (r *ResourceQuotaReconciler) replaceRQCStatus(rqcName string, rqName string, rqNamespace string, status string) {
	reqLogger := r.Log
	rqcFound := &claim.ResourceQuotaClaim{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: rqcName, Namespace: rqNamespace}, rqcFound); err != nil && errors.IsNotFound(err) {
		reqLogger.V(3).Info("ResourceQuotaClaim [ " + rqcName + " ] Not Exists, Do Nothing")
	} else {
		rqcFound.Status.Status = status
		rqcFound.Status.Reason = "ResourceQuota [ " + rqName + " ] Deleted"
		if err := r.Status().Update(context.TODO(), rqcFound); err != nil {
			reqLogger.V(1).Error(err, "Failed to Update ResourceQuotaClaim [ "+rqcName+" ]")
			panic("Failed to Update ResourceQuotaClaim [ " + rqcName + " ]")
		} else {
			reqLogger.V(3).Info("Update ResourceQuotaClaim [ " + rqcName + " ] Success")
		}
	}
}
