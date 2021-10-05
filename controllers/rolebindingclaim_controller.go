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
	"math"
	"math/big"
	"reflect"

	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"

	"crypto/rand"

	"github.com/go-logr/logr"
	rbacApi "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RoleBindingClaimReconciler reconciles a RoleBindingClaim object
type RoleBindingClaimReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
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

	//set helper
	if helper, err := patch.NewHelper(roleBindingClaim, r.Client); err != nil {
		return ctrl.Result{}, err
	} else {
		r.patchHelper = helper
	}
	defer func() {
		r.patchHelper.Patch(context.TODO(), roleBindingClaim)
		// klog.Flush()
	}()

	for idx, _ := range roleBindingClaim.Subjects {
		if roleBindingClaim.Subjects[idx].APIGroup == "" {
			roleBindingClaim.Subjects[idx].APIGroup = "rbac.authorization.k8s.io"
		}
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

	found := &rbacApi.RoleBindingList{}
	//err := r.Get(context.TODO(), types.NamespacedName{Name: roleBindingClaim.ResourceName, Namespace: roleBindingClaim.Namespace}, found)
	labels := map[string]string{"fromClaim": roleBindingClaim.Name}
	err := r.List(context.TODO(), found, client.InNamespace(roleBindingClaim.Namespace), client.MatchingLabels(labels))

	reqLogger.Info("RoleBindingClaim status:" + roleBindingClaim.Status.Status)
	if err != nil {
		reqLogger.Error(err, "Failed to get RoleBinding info")
		return ctrl.Result{}, err
	}

	switch roleBindingClaim.Status.Status {

	case "":
		// Set Owner Annotation from Annotation 'Creator'
		if roleBindingClaim.Annotations != nil && roleBindingClaim.Annotations["creator"] != "" && roleBindingClaim.Annotations["owner"] == "" {
			reqLogger.Info("Set Owner Annotation from Annotation 'Creator'")
			roleBindingClaim.Annotations["owner"] = roleBindingClaim.Annotations["creator"]
		}

		rbcList := &claim.RoleBindingClaimList{}
		if err := r.List(context.TODO(), rbcList, &client.ListOptions{
			Namespace: roleBindingClaim.Namespace,
		}); err != nil {
			reqLogger.Error(err, "Failed to get RoleBindingClaim List")
			panic(err)
		}

		isExistSameName := false
		// 같은 이름의 RBC가 존재하는지 체크
		for _, rbc := range rbcList.Items {
			if (rbc.Status.Status == claim.RoleBindingClaimStatusTypeAwaiting || rbc.Status.Status == claim.RoleBindingClaimStatusTypeSuccess) &&
				rbc.Name == roleBindingClaim.Name {
				isExistSameName = true
				break
			}
		}

		// 같은 resourceName의 RB가 이미 존재하는지 체크
		// if !isExistSameName {
		// 	rbList := &v1.RoleBindingList{}
		// 	if err := r.List(context.TODO(), rbList); err != nil {
		// 		reqLogger.Error(err, "Failed to get RoleBinding List")
		// 		panic(err)
		// 	}

		// 	for _, rb := range rbList.Items {
		// 		if rb.Name == roleBindingClaim.Name && rb.Namespace == roleBindingClaim.Namespace {
		// 			isExistSameName = true
		// 			break
		// 		}
		// 	}
		// }

		if len(found.Items) < 1 && !isExistSameName {
			reqLogger.Info("New RoleBindingClaim Added")
			roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeAwaiting
			roleBindingClaim.Status.Reason = "Please Wait for administrator approval"
		} else {
			reqLogger.Info("RoleBinding [ " + roleBindingClaim.ResourceName + " ] Already Exists.")
			roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeReject
			roleBindingClaim.Status.Reason = "Duplicated RolebindingClaim Name"
		}
	case claim.RoleBindingClaimStatusTypeSuccess:
		rbcLabels := make(map[string]string)
		if roleBindingClaim.Labels != nil {
			rbcLabels = roleBindingClaim.Labels
		}
		rbcLabels["fromClaim"] = roleBindingClaim.Name

		roleBinding := &rbacApi.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:        roleBindingClaim.Name + "-" + MakeRandomHexNumber(),
				Namespace:   roleBindingClaim.Namespace,
				Labels:      rbcLabels,
				Annotations: roleBindingClaim.Annotations,
				Finalizers: []string{
					"rolebinding/finalizers",
				},
			},
			Subjects: roleBindingClaim.Subjects,
			RoleRef:  roleBindingClaim.RoleRef,
		}

		if len(found.Items) < 1 {
			reqLogger.Info("RoleBinding [ " + roleBindingClaim.Name + " ] not Exists, Create RoleBinding.")
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
			reqLogger.Info("RoleBinding [ " + roleBindingClaim.Name + " ] Exists.")

			// if !cmp.Equal(roleBindingClaim.Subjects, found.Items[0].Subjects) || !cmp.Equal(roleBindingClaim.RoleRef, found.Items[0].RoleRef) {
			// 	reqLogger.Info("Same resourceName already exists, modify resourceName and retry.")
			// 	roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeError
			// 	roleBindingClaim.Status.Reason = "Same resourceName already exists, modify resourceName and retry"
			// 	roleBindingClaim.Status.Message = fmt.Errorf("Same resourceName already exists").Error()
			// } else if err := r.Update(context.TODO(), &found.Items[0]); err != nil {
			// 	reqLogger.Error(err, "Failed to update RoleBinding.")
			// 	roleBindingClaim.Status.Status = claim.RoleBindingClaimStatusTypeError
			// 	roleBindingClaim.Status.Reason = "Failed to update RoleBinding"
			// 	roleBindingClaim.Status.Message = err.Error()
			// } else {
			// 	reqLogger.Info("Update RoleBinding Success.")
			// 	roleBindingClaim.Status.Reason = "Update RoleBinding Success"
			// }
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

func MakeRandomHexNumber() string {
	random, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	hexRandom := fmt.Sprintf("%x", random)
	return hexRandom
}
