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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	rbacApi "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"

	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
	"github.com/tmax-cloud/hypercloud-single-operator/util"

	networkv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/yaml"
)

// NamespaceClaimReconciler reconciles a NamespaceClaim object
type NamespaceClaimReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *NamespaceClaimReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log
	reqLogger.Info("Reconciling NamespaceClaim")

	namespaceClaim := &claim.NamespaceClaim{}

	if err := r.Get(context.TODO(), req.NamespacedName, namespaceClaim); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("NamespaceClaim resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get NamespaceClaim")
		return ctrl.Result{}, err
	}

	//set helper
	if helper, err := patch.NewHelper(namespaceClaim, r.Client); err != nil {
		return ctrl.Result{}, err
	} else {
		r.patchHelper = helper
	}
	defer func() {
		r.patchHelper.Patch(context.TODO(), namespaceClaim)
		// klog.Flush()

	}()

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
			namespaceClaim.Status.Status = claim.NamespaceClaimStatueTypeError
			namespaceClaim.Status.Reason = errString
		}
	}()

	nsFound := &v1.Namespace{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: namespaceClaim.ResourceName}, nsFound)

	reqLogger.Info("NamespaceClaim status:" + string(namespaceClaim.Status.Status))
	if err != nil && !errors.IsNotFound(err) {
		reqLogger.Error(err, "Failed to get Namespace info")
		return ctrl.Result{}, err
	}

	switch namespaceClaim.Status.Status {
	case "":
		// Set Owner Annotation from Annotation 'Creator'
		if namespaceClaim.Annotations != nil && namespaceClaim.Annotations["creator"] != "" && namespaceClaim.Annotations["owner"] == "" {
			reqLogger.Info("Set Owner Annotation from Annotation 'Creator'")
			namespaceClaim.Annotations["owner"] = namespaceClaim.Annotations["creator"]
		}

		nscList := &claim.NamespaceClaimList{}
		if err := r.List(context.TODO(), nscList); err != nil {
			reqLogger.Error(err, "Failed to get NamespaceClaim List")
			panic(err)
		}

		flag := false
		for _, nsc := range nscList.Items {
			if nsc.Status.Status == claim.NamespaceClaimStatusTypeAwaiting && nsc.ResourceName == namespaceClaim.ResourceName {
				flag = true
				break
			}
		}

		if err != nil && errors.IsNotFound(err) && !flag {
			reqLogger.Info("Namespace [ " + namespaceClaim.ResourceName + " ] Not found.")
			namespaceClaim.Status.Status = claim.NamespaceClaimStatusTypeAwaiting
			namespaceClaim.Status.Reason = "Please Wait for administrator approval"
		} else {
			reqLogger.Info("Namespace [ " + namespaceClaim.ResourceName + " ] Already Exists.")
			namespaceClaim.Status.Status = claim.NamespaceClaimStatusTypeReject
			namespaceClaim.Status.Reason = "Duplicated NameSpaceName"
		}

	case claim.NamespaceClaimStatusTypeSuccess:
		nscLabels := make(map[string]string)
		if namespaceClaim.Labels != nil {
			nscLabels = namespaceClaim.Labels
		}
		nscLabels["period"] = "1"
		nscLabels["fromClaim"] = namespaceClaim.Name

		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespaceClaim.ResourceName,
				Labels:      nscLabels,
				Annotations: namespaceClaim.Annotations,
				Finalizers: []string{
					"namespace/finalizers",
				},
			},
		}

		resourceQuota := &v1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespaceClaim.ResourceName,
				Namespace:   namespaceClaim.ResourceName,
				Labels:      nscLabels,
				Annotations: namespaceClaim.Annotations,
			},
			Spec: v1.ResourceQuotaSpec{
				//Scopes:        namespaceClaim.Spec.Scopes,
				//ScopeSelector: namespaceClaim.Spec.ScopeSelector,
				Hard: v1.ResourceList{
					v1.ResourceCPU:    namespaceClaim.Spec.Hard["limits.cpu"],
					v1.ResourceMemory: namespaceClaim.Spec.Hard["limits.memory"],
				},
			},
		}

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Create namespace.")
			if err := r.Create(context.TODO(), namespace); err != nil {
				reqLogger.Error(err, "Failed to create Namespace.")
				namespaceClaim.Status.Status = claim.NamespaceClaimStatueTypeError
				namespaceClaim.Status.Reason = "Failed to create Namespace"
				namespaceClaim.Status.Message = err.Error()
			} else if err := r.Create(context.TODO(), resourceQuota); err != nil {
				reqLogger.Error(err, "Failed to create ResourceQuota.")
				namespaceClaim.Status.Status = claim.NamespaceClaimStatueTypeError
				namespaceClaim.Status.Reason = "Failed to create Namespace ResourceQuota"
				namespaceClaim.Status.Message = err.Error()
			} else {
				reqLogger.Info("Create namespace [ " + namespaceClaim.ResourceName + " ] Success")
				//CRB-"ns Name" for nsc user
				r.createCRBForNSCUser(namespaceClaim)

				//Default Network Policy
				r.createDefaultNetPol(namespaceClaim)

				if namespaceClaim.Labels != nil && namespaceClaim.Labels["trial"] != "" && namespaceClaim.Annotations != nil && namespaceClaim.Annotations["owner"] != "" {
					// Trial NamespaceClaim
					r.createTrialRB(namespaceClaim)

					// Set Timers to Send Mail (3 weeks later), Delete Trial NS (1 month later)
					nsResult := &v1.Namespace{}
					if err := r.Get(context.TODO(), types.NamespacedName{Name: namespaceClaim.ResourceName}, nsResult); err != nil {
						reqLogger.Error(err, "Failed to Read Namespace  [ "+namespaceClaim.ResourceName+" ]")
						panic("Failed to Read Namespace  [ " + namespaceClaim.ResourceName + " ]")
					}
					util.SetTrialNSTimer(nsResult, r.Client, reqLogger)

					// Send Success Confirm Mail //TODO
					r.sendConfirmMail(namespaceClaim, nsResult.CreationTimestamp.Time, true, reqLogger)
				} else {
					// Make Namespaced RoleBinding for non-trial User
					r.createNSCRoleBinding(namespaceClaim)
				}

				namespaceClaim.Status.Reason = "Create Namespace Success"

				httpgrafanaurl := "https://" + util.HYPERCLOUD_API_SERVER_URI + "grafanaDashboard"
				GrafanaBody := `{
					"email": "` + namespaceClaim.Annotations["owner"] + `",
					"namespace": "` + namespaceClaim.ResourceName + `"
				}`
				reqLogger.Info("Request Body : " + GrafanaBody)
				request, _ := http.NewRequest("POST", httpgrafanaurl, bytes.NewBuffer([]byte(GrafanaBody)))
				http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
				client := &http.Client{}
				resp, err := client.Do(request)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
			}
		} else {
			reqLogger.Info("Update namespace.")
			if err := r.Update(context.TODO(), namespace); err != nil {
				reqLogger.Error(err, "Failed to update Namespace.")
				namespaceClaim.Status.Status = claim.NamespaceClaimStatueTypeError
				namespaceClaim.Status.Reason = "Failed to update Namespace"
				namespaceClaim.Status.Message = err.Error()
			} else if err := r.Update(context.TODO(), resourceQuota); err != nil {
				reqLogger.Error(err, "Failed to update ResourceQuota.")
				namespaceClaim.Status.Status = claim.NamespaceClaimStatueTypeError
				namespaceClaim.Status.Reason = "Failed to update Namespace ResourceQuota"
				namespaceClaim.Status.Message = err.Error()
			} else {
				reqLogger.Info("Update Namespace Success")
				namespaceClaim.Status.Reason = "Update Namespace Success"
			}
		}
	case claim.NamespaceClaimStatusTypeReject:
		if namespaceClaim.Labels != nil && namespaceClaim.Labels["trial"] != "" && namespaceClaim.Annotations != nil && namespaceClaim.Annotations["owner"] != "" && namespaceClaim.Status.Message != "reject mail sent" {
			r.sendConfirmMail(namespaceClaim, time.Now(), false, reqLogger)
			namespaceClaim.Status.Message = "reject mail sent"
		}
	}
	namespaceClaim.Status.LastTransitionTime = metav1.Now()
	return ctrl.Result{}, nil
}

func (r *NamespaceClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&claim.NamespaceClaim{}).
		WithEventFilter(
			predicate.Funcs{
				// Only reconciling if the status.status change
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldRbcStatus := e.ObjectOld.(*claim.NamespaceClaim).DeepCopy().Status.Status
					newRbcStatus := e.ObjectNew.(*claim.NamespaceClaim).DeepCopy().Status.Status
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

func (r *NamespaceClaimReconciler) sendConfirmMail(namespaceClaim *claim.NamespaceClaim, createTime time.Time, isSuccess bool, reqLogger logr.Logger) {
	var subject string
	var body string
	var imgPath string
	var imgCid string
	email := namespaceClaim.Annotations["owner"]
	if isSuccess {
		subject = "HyperCloud 서비스 신청 승인 완료"
		body = util.TRIAL_SUCCESS_CONFIRM_MAIL_CONTENTS
		body = strings.ReplaceAll(body, "%%NAMESPACE_NAME%%", namespaceClaim.ResourceName)
		body = strings.ReplaceAll(body, "%%TRIAL_START_TIME%%", createTime.Format("2006-01-02"))
		body = strings.ReplaceAll(body, "%%TRIAL_END_TIME%%", createTime.AddDate(0, 0, 30).Format("2006-01-02"))
		imgPath = "/img/trial-approval.png"
		imgCid = "trial-approval"

	} else {
		subject = "HyperCloud 서비스 신청 승인 거절"
		body = util.TRIAL_FAIL_CONFIRM_MAIL_CONTENTS
		if namespaceClaim.Status.Reason != "" {
			body = strings.ReplaceAll(body, "%%FAIL_REASON%%", namespaceClaim.Status.Reason)
		} else {
			body = strings.ReplaceAll(body, "%%FAIL_REASON%%", "Unknown Reason")
		}
		imgPath = "/img/trial-disapproval.png"
		imgCid = "trial-disapproval"
	}
	util.SendMail(email, subject, body, imgPath, imgCid, reqLogger)
}

func (r *NamespaceClaimReconciler) createNSCRoleBinding(namespaceClaim *claim.NamespaceClaim) {
	reqLogger := r.Log
	reqLogger.Info("Create RoleBinding For NamespaceClaim user Start")
	rbForNscUserFound := &rbacApi.RoleBinding{}

	if err := r.Get(context.TODO(), types.NamespacedName{Name: namespaceClaim.ResourceName + "-ns-listget", Namespace: namespaceClaim.ResourceName}, rbForNscUserFound); err != nil && errors.IsNotFound(err) {
		rbForNscUser := &rbacApi.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespaceClaim.ResourceName + "-ns-listget",
				Namespace:   namespaceClaim.ResourceName,
				Labels:      namespaceClaim.Labels,
				Annotations: namespaceClaim.Annotations,
			},
			Subjects: []rbacApi.Subject{
				{
					Kind:     "User",
					APIGroup: util.RBAC_API_GROUP,
					Name:     namespaceClaim.Annotations["owner"],
				},
			},
			RoleRef: rbacApi.RoleRef{
				Kind:     "ClusterRole",
				APIGroup: util.RBAC_API_GROUP,
				Name:     "namespace-owner",
			},
		}
		if err := r.Create(context.TODO(), rbForNscUser); err != nil && errors.IsNotFound(err) {
			reqLogger.Info("RoleBinding for NameSpace [ " + namespaceClaim.ResourceName + "-ns-listget ] Already Exists")
		} else {
			reqLogger.Info("Create RoleBinding [ " + namespaceClaim.ResourceName + "-ns-listget ] Success")
		}
	} else {
		reqLogger.Info("RoleBinding for NameSpace [ " + namespaceClaim.ResourceName + " ] Already Exists")
	}
}

func (r *NamespaceClaimReconciler) createTrialRB(namespaceClaim *claim.NamespaceClaim) {
	reqLogger := r.Log
	reqLogger.Info("Create Trial-RoleBinding For NamespaceClaim user Start")
	rbForTrialUserFound := &rbacApi.RoleBinding{}

	if err := r.Get(context.TODO(), types.NamespacedName{Name: "trial-" + namespaceClaim.ResourceName, Namespace: namespaceClaim.ResourceName}, rbForTrialUserFound); err != nil && errors.IsNotFound(err) {
		rbForTrialUser := &rbacApi.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "trial-" + namespaceClaim.ResourceName,
				Namespace:   namespaceClaim.ResourceName,
				Labels:      namespaceClaim.Labels,
				Annotations: namespaceClaim.Annotations,
			},
			Subjects: []rbacApi.Subject{
				{
					Kind:     "User",
					APIGroup: util.RBAC_API_GROUP,
					Name:     namespaceClaim.Annotations["owner"],
				},
			},
			RoleRef: rbacApi.RoleRef{
				Kind:     "ClusterRole",
				APIGroup: util.RBAC_API_GROUP,
				Name:     "namespace-owner",
			},
		}
		if err := r.Create(context.TODO(), rbForTrialUser); err != nil && errors.IsNotFound(err) {
			reqLogger.Info("RoleBinding for Trial NameSpace [ trial-" + namespaceClaim.ResourceName + " ] Already Exists")
		} else {
			reqLogger.Info("Create RoleBinding [ trial-" + namespaceClaim.ResourceName + " ] Success")
		}
	} else {
		reqLogger.Info("TrialRoleBinding for Trial NameSpace [ " + namespaceClaim.ResourceName + " ] Already Exists")
	}
}

func (r *NamespaceClaimReconciler) createDefaultNetPol(namespaceClaim *claim.NamespaceClaim) {
	reqLogger := r.Log
	reqLogger.Info("Create Network Policy for new Namespace [ " + namespaceClaim.ResourceName + " ] Start")
	netPolConfigFound := &v1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: util.DEFAULT_NETWORK_POLICY_CONFIG_MAP, Namespace: util.HYPERCLOUD_NAMESPACE}, netPolConfigFound); err != nil && errors.IsNotFound(err) {
		// Make ConfigMap default-networkpolicy-configmap With Empty Data in hypercloud5-system Namespace
		netPolConfig := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      util.DEFAULT_NETWORK_POLICY_CONFIG_MAP,
				Namespace: util.HYPERCLOUD_NAMESPACE,
			},
			Data: map[string]string{
				util.NETWORK_POLICY_YAML: "",
			},
		}
		if err := r.Create(context.TODO(), netPolConfig); err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap  [ "+util.DEFAULT_NETWORK_POLICY_CONFIG_MAP+" ]")
			panic("Failed to create ConfigMap  [ " + util.DEFAULT_NETWORK_POLICY_CONFIG_MAP + " ]")
		} else {
			reqLogger.Info("Create ConfigMap  [ " + util.DEFAULT_NETWORK_POLICY_CONFIG_MAP + " ] Success")
		}

	} else {
		// Read ConfigMap & Create Network Policy If Data Exists
		if netPolConfigFound != nil && netPolConfigFound.Data != nil && netPolConfigFound.Data[util.NETWORK_POLICY_YAML] != "" {
			netPolYamlString := netPolConfigFound.Data[util.NETWORK_POLICY_YAML]
			reqLogger.Info(" netPolYamlString : " + netPolYamlString)
			reqLogger.Info(" ------------------------------------------------------ ")
			netPol := &networkv1.NetworkPolicy{}
			if err := yaml.Unmarshal([]byte(netPolYamlString), &netPol); err != nil {
				panic("Failed to Convert yaml to json  [ " + util.NETWORK_POLICY_YAML + " ]")
			}
			netPol.SetName(namespaceClaim.ResourceName)
			netPol.SetNamespace(namespaceClaim.ResourceName)
			if err := r.Create(context.TODO(), netPol); err != nil {
				reqLogger.Error(err, "Failed to create NetworkPolicy  [ "+namespaceClaim.ResourceName+" ]")
				panic("Failed to create NetworkPolicy  [ " + namespaceClaim.ResourceName + " ]")
			} else {
				reqLogger.Info("Create NetworkPolicy  [ " + namespaceClaim.ResourceName + " ] Success")
			}
		}
	}

}

func (r *NamespaceClaimReconciler) createCRBForNSCUser(namespaceClaim *claim.NamespaceClaim) {
	reqLogger := r.Log
	reqLogger.Info("Create ClusterRoleBinding For NamespaceClaim user Start")
	crbForNscUserFound := &rbacApi.ClusterRoleBinding{}
	crbForNscUser := &rbacApi.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "CRB-" + namespaceClaim.ResourceName,
			Labels:      namespaceClaim.Labels,
			Annotations: namespaceClaim.Annotations,
		},
		Subjects: []rbacApi.Subject{
			{
				Kind:     "User",
				APIGroup: util.RBAC_API_GROUP,
				Name:     namespaceClaim.Annotations["owner"],
			},
		},
		RoleRef: rbacApi.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: util.RBAC_API_GROUP,
			Name:     "clusterrole-trial",
		},
	}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: "CRB-" + namespaceClaim.ResourceName}, crbForNscUserFound); err != nil && errors.IsNotFound(err) {
		if err := r.Create(context.TODO(), crbForNscUser); err != nil {
			reqLogger.Error(err, "Failed to create ClusterRoleBinding [ CRB-"+namespaceClaim.ResourceName+" ]")
			panic("Failed to create ClusterRoleBinding [ CRB-" + namespaceClaim.ResourceName + " ]")
		} else {
			reqLogger.Info("Create ClusterRoleBinding [ CRB-" + namespaceClaim.ResourceName + " ] Success")
		}
	} else {
		if err := r.Delete(context.TODO(), crbForNscUserFound); err != nil {
			reqLogger.Error(err, "Failed to Delete ClusterRoleBinding [ CRB-"+namespaceClaim.ResourceName+" ]")
			panic("Failed to update ClusterRoleBinding [ CRB-" + namespaceClaim.ResourceName + " ]")
		} else if err := r.Create(context.TODO(), crbForNscUser); err != nil {
			reqLogger.Error(err, "Failed to Re-Create ClusterRoleBinding [ CRB-"+namespaceClaim.ResourceName+" ]")
			panic("Failed to update ClusterRoleBinding [ CRB-" + namespaceClaim.ResourceName + " ]")
		} else {
			reqLogger.Info("Update ClusterRoleBinding [ CRB-" + namespaceClaim.ResourceName + " ] Success")
		}
	}
}
