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
	"reflect"

	"github.com/tmax-cloud/hypercloud-go-operator/util"

	"fmt"

	claim "github.com/tmax-cloud/hypercloud-go-operator/apis/claim/v1alpha1"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
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

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log
	// your logic here
	reqLogger.Info("Reconciling Namespace")
	namespace := &v1.Namespace{}

	if err := r.Get(context.TODO(), req.NamespacedName, namespace); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Namespace resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		reqLogger.Error(err, "Failed to get Namespace")
		return ctrl.Result{}, err
	}

	//set helper
	if helper, err := patch.NewHelper(namespace, r.Client); err != nil {
		return ctrl.Result{}, err
	} else {
		r.patchHelper = helper
	}
	defer func() {
		r.patchHelper.Patch(context.TODO(), namespace)

	}()

	defer func() {
		s := recover()
		if s != nil {
			fmt.Println("Error !! : ", s)
			// var errString string
			// switch x := s.(type) {
			// case string:
			// 	errString = x
			// case error:
			// 	errString = x.Error()
			// default:
			// 	errString = "unknown error"
			// }
		}
	}()

	switch namespace.Status.Phase {

	case "Terminating":
		if namespace.Labels != nil && namespace.Labels["fromClaim"] != "" {
			reqLogger.Info("Namespace from Claim [ " + namespace.Name + " ] is in Terminating Status")
			if namespace.Finalizers != nil {
				namespace.Finalizers = util.RemoveValue(namespace.Finalizers, "namespace/finalizers")
			}
			reqLogger.Info("Delete Finalizer [ namespace/finalizers ] Success")

			// Delete ClusterRoleBinding for nsc user
			r.deleteCRBForNSCUser(namespace)

			reqLogger.Info("Update NamespaceClaim [ " + namespace.Labels["fromClaim"] + " ] Status to Deleted")
			r.replaceNSCStatus(namespace.Labels["fromClaim"], claim.NamespaceClaimStatueTypeDeleted)

		}
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Namespace{}).
		WithEventFilter(
			predicate.Funcs{
				// Only reconciling if the status.status change
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldRbcStatus := e.ObjectOld.(*v1.Namespace).DeepCopy().Status.Phase
					newRbcStatus := e.ObjectNew.(*v1.Namespace).DeepCopy().Status.Phase
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

func (r *NamespaceReconciler) replaceNSCStatus(nscName string, status string) {
	reqLogger := r.Log
	reqLogger.Info("Replace NamespaceClaim [ " + nscName + " ] Status to " + status)
	nscFound := &claim.NamespaceClaim{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: nscName}, nscFound); err != nil && errors.IsNotFound(err) {
		reqLogger.Info("NamespaceClaim [ " + nscName + " ] Not Exists, Do Nothing")
	} else {
		nscFound.Status.Status = status
		if err := r.Update(context.TODO(), nscFound); err != nil {
			reqLogger.Error(err, "Failed to Update ClusterRoleBinding [ "+nscName+" ]")
			panic("Failed to Update ClusterRoleBinding [ " + nscName + " ]")
		} else {
			reqLogger.Info("Update ClusterRoleBinding [ " + nscName + " ] Success")
		}
	}
}

func (r *NamespaceReconciler) deleteCRBForNSCUser(namespace *v1.Namespace) {
	reqLogger := r.Log
	reqLogger.Info("Delete ClusterRoleBinding For NamespaceClaim user Start")
	crbForNscUserFound := &rbacApi.ClusterRoleBinding{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: "CRB-" + namespace.Name}, crbForNscUserFound); err != nil && errors.IsNotFound(err) {
		reqLogger.Info("ClusterRoleBinding [ CRB-" + namespace.Name + " ] Not Exists, Do Nothing")
	} else {
		if err := r.Delete(context.TODO(), crbForNscUserFound); err != nil {
			reqLogger.Error(err, "Failed to Delete ClusterRoleBinding [ CRB-"+namespace.Name+" ]")
			panic("Failed to Delete ClusterRoleBinding [ CRB-" + namespace.Name + " ]")
		} else {
			reqLogger.Info("Delete ClusterRoleBinding [ CRB-" + namespace.Name + " ] Success")
		}
	}
}
