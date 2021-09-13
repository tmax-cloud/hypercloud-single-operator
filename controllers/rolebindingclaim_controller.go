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
	"reflect"

	"github.com/google/go-cmp/cmp"
	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"

	"github.com/go-logr/logr"
	rbacApi "k8s.io/api/rbac/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RoleBindingClaimReconciler reconciles a RoleBindingClaim object
type RoleBindingClaimReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *RoleBindingClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log

	// your logic here
	reqLogger.Info("Reconciling RoleBindingClaim")
	roleBindingClaim := &claim.RoleBindingClaim{}

	if err := r.Get(context.TODO(), req.NamespacedName, roleBindingClaim); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("RoleBindingClaim resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		reqLogger.Error(err, "Failed to get RoleBindingClaim")
		return ctrl.Result{}, err
	}

	defer func() {
		s := recover()
		if s != nil {
			fmt.Println("Error !! : ", s)
			var errString string
			switch x := s.(type) {
			case string:
				errString = x
			case error:
				errString = x.Error()
			default:
				errString = "unknown error"
			}
			roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeError
			roleBindingClaim.Status.Reason = errString
		}
	}()

	found := &rbacApi.RoleBinding{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: roleBindingClaim.ResourceName, Namespace: roleBindingClaim.Namespace}, found)

	reqLogger.Info("RoleBindingClaim status:" + roleBindingClaim.Status.Status)
	if err != nil && !errors.IsNotFound(err) {
		reqLogger.Error(err, "Failed to get RoleBinding info")
		return ctrl.Result{}, err
	}

	switch roleBindingClaim.Status.Status {

	case "":
		rbcList := &claim.RoleBindingClaimList{}
		if err := r.List(context.TODO(), rbcList); err != nil {
			reqLogger.Error(err, "Failed to get RoleBindingClaim List")
			panic(err)
		}

		flag := false
		for _, rbc := range rbcList.Items {
			if (rbc.Status.Status == claim.RoleBindingClaimStatusTypeAwaiting || rbc.Status.Status == claim.RoleBindingClaimStatusTypeSuccess) &&
				rbc.Namespace == roleBindingClaim.Namespace &&
				rbc.ResourceName == roleBindingClaim.ResourceName {
				flag = true
				break
			}
		}

		if !flag {
			rbList := &v1.RoleBindingList{}
			if err := r.List(context.TODO(), rbList); err != nil {
				reqLogger.Error(err, "Failed to get RoleBinding List")
				panic(err)
			}

			for _, rb := range rbList.Items {
				if rb.Name == roleBindingClaim.Name && rb.Namespace == roleBindingClaim.Namespace {
					flag = true
					break
				}
			}
		}

		if err != nil && errors.IsNotFound(err) && !flag {
			reqLogger.Info("New RoleBindingClaim Added")
			roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeAwaiting
			roleBindingClaim.Status.Reason = "Please Wait for administrator approval"
		} else {
			reqLogger.Info("Namespace [ " + roleBindingClaim.ResourceName + " ] Already Exists.")
			roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeReject
			roleBindingClaim.Status.Reason = "Duplicated Rolebinding Name"
		}
	case claim.RoleBindingClaimStatusTypeSuccess:
		rbcLabels := make(map[string]string)
		if roleBindingClaim.Labels != nil {
			rbcLabels = roleBindingClaim.Labels
		}
		rbcLabels["fromClaim"] = roleBindingClaim.Name

		roleBinding := &rbacApi.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleBindingClaim.ResourceName,
				Namespace: roleBindingClaim.Namespace,
				Labels:    rbcLabels,
				Finalizers: []string{
					"rolebinding/finalizers",
				},
			},
			Subjects: roleBindingClaim.Subjects,
			RoleRef:  roleBindingClaim.RoleRef,
		}

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("RoleBinding [ " + roleBindingClaim.ResourceName + " ] not Exists, Create RoleBinding.")
			if err := r.Create(context.TODO(), roleBinding); err != nil {
				reqLogger.Error(err, "Failed to create RoleBinding.")
				roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeError
				roleBindingClaim.Status.Reason = "Failed to create RoleBinding"
				roleBindingClaim.Status.Message = err.Error()
			} else {
				reqLogger.Info("Create RoleBinding Success.")
				roleBindingClaim.Status.Reason = "Create RoleBinding Success"
			}
		} else {
			reqLogger.Info("RoleBinding [ " + roleBindingClaim.ResourceName + " ] Exists, Update RoleBinding.")

			if cmp.Equal(roleBindingClaim.Subjects, roleBinding.Subjects) {
				reqLogger.Info("Same resourceName already exists, modify resourceName and retry.")
				roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeError
				roleBindingClaim.Status.Reason = "Same resourceName already exists, modify resourceName and retry"
				roleBindingClaim.Status.Message = fmt.Errorf("Same resourceName already exists").Error()
			} else if err := r.Update(context.TODO(), roleBinding); err != nil {
				reqLogger.Error(err, "Failed to update RoleBinding.")
				roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeError
				roleBindingClaim.Status.Reason = "Failed to update RoleBinding"
				roleBindingClaim.Status.Message = err.Error()
			} else {
				reqLogger.Info("Update RoleBinding Success.")
				roleBindingClaim.Status.Reason = "Update RoleBinding Success"
			}
		}
	}

	roleBindingClaim.Status.LastTransitionTime = metav1.Now()
	if err := r.Status().Update(context.TODO(), roleBindingClaim); err != nil {
		reqLogger.Error(err, "Failed to update roleBindingClaim status.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RoleBindingClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claim.RoleBindingClaim{}).
		WithEventFilter(
			predicate.Funcs{
				// Only reconciling if the status.status change
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldRbcStatus := e.ObjectOld.(*claim.RoleBindingClaim).DeepCopy().Status.Status
					newRbcStatus := e.ObjectNew.(*claim.RoleBindingClaim).DeepCopy().Status.Status
					if !reflect.DeepEqual(oldRbcStatus, newRbcStatus) {
						return true
					} else {
						return false
					}
				},
			},
		).
		Complete(r)
}
