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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	// "k8s.io/kubernetes/pkg/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
)

// ResourceQuotaClaimReconciler reconciles a ResourceQuotaClaim object
type ResourceQuotaClaimReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *ResourceQuotaClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log
	// your logic here
	reqLogger.Info("Reconciling ResourceQuotaClaim")
	resourceQuotaClaim := &claim.ResourceQuotaClaim{}

	if err := r.Get(context.TODO(), req.NamespacedName, resourceQuotaClaim); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("ResourceQuotaClaim resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}

		reqLogger.Error(err, "Failed to get ResourceQuotaClaim")
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
			resourceQuotaClaim.Status.Status = claim.ResourceQuotaClaimStatusTypeError
			resourceQuotaClaim.Status.Reason = errString
		}
	}()

	found := &v1.ResourceQuota{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: resourceQuotaClaim.ResourceName, Namespace: resourceQuotaClaim.Namespace}, found)

	reqLogger.Info("ResourceQuotaClaim status:" + resourceQuotaClaim.Status.Status)
	if err != nil && !errors.IsNotFound(err) {
		reqLogger.Error(err, "failed to get ResourceQuota info")
		return ctrl.Result{}, err
	}

	switch resourceQuotaClaim.Status.Status {

	case "":
		rqcList := &claim.ResourceQuotaClaimList{}
		if err := r.List(context.TODO(), rqcList); err != nil {
			reqLogger.Error(err, "Failed to get ResourceQuotaClaim List")
			panic(err)
		}

		flag := false
		for _, rqc := range rqcList.Items {
			if (rqc.Status.Status == claim.ResourceQuotaClaimStatusTypeAwaiting || rqc.Status.Status == claim.ResourceQuotaClaimStatusTypeSuccess) &&
				rqc.Namespace == resourceQuotaClaim.Namespace &&
				rqc.ResourceName == resourceQuotaClaim.ResourceName {
				flag = true
				break
			}
		}

		if err != nil && errors.IsNotFound(err) && !flag {
			reqLogger.Info("New ResourceQuotaClaim Added")
			resourceQuotaClaim.Status.Status = claim.ResourceQuotaClaimStatusTypeAwaiting
			resourceQuotaClaim.Status.Reason = "Please Wait for administrator approval"
		} else {
			reqLogger.Info("Namespace [ " + resourceQuotaClaim.ResourceName + " ] Already Exists.")
			resourceQuotaClaim.Status.Status = claim.ResourceQuotaClaimStatusTypeReject
			resourceQuotaClaim.Status.Reason = "Duplicated ResourceQuota Name"
		}
	case claim.ResourceQuotaClaimStatusTypeSuccess:
		// hards := v1.ResourceList{

		// 	v1.ResourceLimitsCPU.String():    resourceQuotaClaim.SpecLimit.LimitCpu,
		// 	v1.ResourceLimitsMemory.String(): resourceQuotaClaim.SpecLimit.LimitCpu,
		// }

		rqcLabels := make(map[string]string)
		if resourceQuotaClaim.Labels != nil {
			rqcLabels = resourceQuotaClaim.Labels
		}
		rqcLabels["fromClaim"] = resourceQuotaClaim.Name

		resourceQuota := &v1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceQuotaClaim.ResourceName,
				Namespace: resourceQuotaClaim.Namespace,
				Labels:    rqcLabels,
				Finalizers: []string{
					"resourcequota/finalizers",
				},
			},
			// Spec: v1.ResourceQuotaSpec{
			// 	//Scopes:        resourceQuotaClaim.Spec.Scopes,
			// 	//ScopeSelector: resourceQuotaClaim.Spec.ScopeSelector,
			// 	Hard: v1.ResourceList{
			// 		v1.ResourceCPU:    resourceQuotaClaim.Spec.Hard["limits.cpu"],
			// 		v1.ResourceMemory: resourceQuotaClaim.Spec.Hard["limits.memory"],
			// 	},
			// },
		}

		hardList := make(map[v1.ResourceName]resource.Quantity)
		hardList[v1.ResourceLimitsCPU] = resourceQuotaClaim.Spec.Hard["limits.cpu"]
		hardList[v1.ResourceLimitsMemory] = resourceQuotaClaim.Spec.Hard["limits.memory"]

		if (resourceQuotaClaim.Spec.Hard["requests.cpu"] != resource.Quantity{}) {
			hardList[v1.ResourceRequestsCPU] = resourceQuotaClaim.Spec.Hard["requests.cpu"]
		} else if (resourceQuotaClaim.Spec.Hard["cpu"] != resource.Quantity{}) {
			hardList[v1.ResourceRequestsCPU] = resourceQuotaClaim.Spec.Hard["cpu"]
		}

		if (resourceQuotaClaim.Spec.Hard["requests.memory"] != resource.Quantity{}) {
			hardList[v1.ResourceRequestsMemory] = resourceQuotaClaim.Spec.Hard["requests.memory"]
		} else if (resourceQuotaClaim.Spec.Hard["memory"] != resource.Quantity{}) {
			hardList[v1.ResourceRequestsMemory] = resourceQuotaClaim.Spec.Hard["memory"]
		}

		resourceQuota.Spec.Hard = hardList

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("ResourceQuota [ " + resourceQuotaClaim.ResourceName + " ] not Exists, Create ResourceQuota.")
			if err := r.Create(context.TODO(), resourceQuota); err != nil {
				reqLogger.Error(err, "Failed to create ResourceQuota.")
				resourceQuotaClaim.Status.Status = claim.ResourceQuotaClaimStatusTypeError
				resourceQuotaClaim.Status.Reason = "Failed to create ResourceQuota"
				resourceQuotaClaim.Status.Message = err.Error()
			} else {
				reqLogger.Info("Create ResourceQuota Success.")
				resourceQuotaClaim.Status.Reason = "Create ResourceQuota Success"
			}
		} else {
			reqLogger.Info("ResourceQuota [ " + resourceQuotaClaim.ResourceName + " ] Exists, Update ResourceQuota.")
			if err := r.Update(context.TODO(), resourceQuota); err != nil {
				reqLogger.Error(err, "Failed to update ResourceQuota.")
				resourceQuotaClaim.Status.Status = claim.ResourceQuotaClaimStatusTypeError
				resourceQuotaClaim.Status.Reason = "Failed to update ResourceQuota"
				resourceQuotaClaim.Status.Message = err.Error()
			} else {
				reqLogger.Info("Update ResourceQuota Success.")
				resourceQuotaClaim.Status.Reason = "Update ResourceQuota Success"
			}
		}
	}

	resourceQuotaClaim.Status.LastTransitionTime = metav1.Now()
	if err := r.Status().Update(context.TODO(), resourceQuotaClaim); err != nil {
		reqLogger.Error(err, "Failed to update ResourceQuotaClaim status.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ResourceQuotaClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claim.ResourceQuotaClaim{}).
		WithEventFilter(
			predicate.Funcs{
				// Only reconciling if the status.status change
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldRqcStatus := e.ObjectOld.(*claim.ResourceQuotaClaim).DeepCopy().Status.Status
					newRqcStatus := e.ObjectNew.(*claim.ResourceQuotaClaim).DeepCopy().Status.Status
					if !reflect.DeepEqual(oldRqcStatus, newRqcStatus) {
						return true
					} else {
						return false
					}
				},
			},
		).
		Complete(r)
}
