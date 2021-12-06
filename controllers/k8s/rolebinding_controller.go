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
	rbacApi "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RoleBindingReconciler reconciles a RoleBinding object
type RoleBindingReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
}

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io.tmax.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io.tmax.io,resources=rolebindings/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RoleBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *RoleBindingReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log

	reqLogger.Info("Reconciling RoleBinding")
	rolebinding := &rbacApi.RoleBinding{}

	if err := r.Get(context.TODO(), req.NamespacedName, rolebinding); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("RoleBinding resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		} else {
			reqLogger.Error(err, "Failed to get RoleBinding")
			return ctrl.Result{}, err
		}
	}

	//set helper
	if helper, err := patch.NewHelper(rolebinding, r.Client); err != nil {
		return ctrl.Result{}, err
	} else {
		r.patchHelper = helper
	}
	defer func() {
		r.patchHelper.Patch(context.TODO(), rolebinding)
	}()

	defer func() {
		s := recover()
		if s != nil {
			fmt.Println("Error !! : ", s)
		}
	}()

	if rolebinding.DeletionTimestamp != nil {
		if rolebinding.Labels != nil && rolebinding.Labels["fromClaim"] != "" {
			if rolebinding.Finalizers != nil {
				rolebinding.Finalizers = util.RemoveValue(rolebinding.Finalizers, "rolebinding/finalizers")
				reqLogger.Info("Delete Finalizer [ rolebinding/finalizers ] Success")
			}

			r.replaceRBCStatus(rolebinding.Labels["fromClaim"], rolebinding.Name, rolebinding.Namespace, claim.RoleBindingClaimStatusTypeDeleted)
			reqLogger.Info("Update RoleBindingClaim [ " + rolebinding.Labels["fromClaim"] + " ] Status to RoleBinding Deleted")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacApi.RoleBinding{}).
		WithEventFilter(
			predicate.Funcs{
				// Reconciling only when RoleBinding is deleted,
				// which is noticed by DeletionTimestamp
				DeleteFunc: func(e event.DeleteEvent) bool {
					return false
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					newRB := e.ObjectNew.(*rbacApi.RoleBinding).DeepCopy()
					if newRB.DeletionTimestamp != nil {
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

func (r *RoleBindingReconciler) replaceRBCStatus(rbcName string, rbName string, rbNamespace string, status string) {
	reqLogger := r.Log
	rbcFound := &claim.RoleBindingClaim{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: rbcName, Namespace: rbNamespace}, rbcFound); err != nil && errors.IsNotFound(err) {
		reqLogger.Info("RoleBindingClaim [ " + rbcName + " ] Not Exists, Do Nothing")
	} else {
		rbcFound.Status.Status = status
		rbcFound.Status.Reason = "RoleBinding [ " + rbName + " ] Deleted"
		if err := r.Status().Update(context.TODO(), rbcFound); err != nil {
			reqLogger.Error(err, "Failed to Update RoleBindingClaim [ "+rbcName+" ]")
			panic("Failed to Update RoleBindingClaim [ " + rbcName + " ]")
		} else {
			reqLogger.Info("Update RoleBindingClaim [ " + rbcName + " ] Success")
		}
	}
}
