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
	"encoding/json"
	goError "errors"
	"io/ioutil"
	"net/http"

	"github.com/tmax-cloud/hypercloud-single-operator/util"

	"fmt"

	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"

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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	patchHelper *patch.Helper
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

var url string

func init() {
	url = "https://" + util.HYPERCLOUD_API_SERVER_URI + "namespace"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (r *NamespaceReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if namespace.Status.Phase == "Terminating" {
		reqLogger.Info(namespace.Name + " is in Terminating Status")
		if namespace.Labels != nil && namespace.Labels["fromClaim"] != "" {
			if namespace.Finalizers != nil {
				namespace.Finalizers = util.RemoveValue(namespace.Finalizers, "namespace/finalizers")
			}
			reqLogger.Info("Delete Finalizer [ namespace/finalizers ] Success")

			// Delete ClusterRoleBinding for nsc user
			r.deleteCRBForNSCUser(namespace)

			reqLogger.Info("Update NamespaceClaim [ " + namespace.Labels["fromClaim"] + " ] Status to Namespace Deleted")
			r.replaceNSCStatus(namespace.Labels["fromClaim"], namespace.Name, claim.NamespaceClaimStatusTypeDeleted)
		}
		// if namespace is deleted
		// request broadcast to hypercloud-api-server
		response, err := r.postRequestToHypercloudApiServer(namespace, "DELETED")
		if err != nil {
			reqLogger.Error(err, " Failed to broadcast namespace delete event")
		}
		reqLogger.Info(response)

		httpgrafanaurl := "https://" + util.HYPERCLOUD_API_SERVER_URI + "grafanaDashboard?namespace=" + namespace.Name
		request, _ := http.NewRequest("DELETE", httpgrafanaurl, nil)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			reqLogger.Error(err, "Grafana Failed")
		} else {
			defer resp.Body.Close()
		}
	} else {
		// if namespace is modified
		// request broadcast to hypercloud-api-server
		response, err := r.postRequestToHypercloudApiServer(namespace, "MODIFIED")
		if err != nil {
			reqLogger.Error(err, "Failed to broadcast namespace modify event")
		}
		reqLogger.Info(response)
	}

	if namespace.Labels != nil && namespace.Labels["trial"] != "" && namespace.Labels["period"] != "" && namespace.Annotations["owner"] != "" {
		util.SetTrialNSTimer(namespace, r.Client, reqLogger)
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controller, err := ctrl.NewControllerManagedBy(mgr).
		For(&v1.Namespace{}).
		WithEventFilter(
			predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					// Only reconciling if the status.status change
					// oldNsStatus := e.ObjectOld.(*v1.Namespace).DeepCopy().Status.Phase
					// newNsStatus := e.ObjectNew.(*v1.Namespace).DeepCopy().Status.Phase

					// oldNsLabels := e.ObjectOld.(*v1.Namespace).DeepCopy().Labels
					// newNsLabels := e.ObjectNew.(*v1.Namespace).DeepCopy().Labels

					// if !reflect.DeepEqual(oldNsStatus, newNsStatus) || !reflect.DeepEqual(oldNsLabels, newNsLabels) {
					// 	return true
					// } else {
					// 	return false
					// }
					return true
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
			},
		).
		Build(r)

	if err != nil {
		return err
	}

	return controller.Watch(
		&source.Kind{Type: &v1.Namespace{}},
		handler.EnqueueRequestsFromMapFunc(r.reconcileNamespaceForCreateEvent),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		},
	)
}

func (r *NamespaceReconciler) reconcileNamespaceForCreateEvent(o client.Object) []ctrl.Request {
	reqLogger := r.Log

	ns_name := o.GetName()
	ns_namespace := o.GetNamespace()
	namespace := &v1.Namespace{}

	if err := r.Get(context.TODO(), types.NamespacedName{Name: ns_name, Namespace: ns_namespace}, namespace); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Namespace resource not found. Ignoring since object must be deleted.")
			return nil
		}
		reqLogger.Error(err, "Failed to get Namespace")
		return nil
	}

	// if namespace is created
	// request broadcast to hypercloud-api-server
	response, err := r.postRequestToHypercloudApiServer(namespace, "ADDED")
	if err != nil {
		reqLogger.Error(err, "Failed to broadcast namespace create event")
		return nil
	}
	reqLogger.Info(response)
	return nil
}

func (r *NamespaceReconciler) replaceNSCStatus(nscName string, nsName string, status string) {
	reqLogger := r.Log
	nscFound := &claim.NamespaceClaim{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: nscName}, nscFound); err != nil && errors.IsNotFound(err) {
		reqLogger.Info("NamespaceClaim [ " + nscName + " ] Not Exists, Do Nothing")
	} else {
		nscFound.Status.Status = status
		nscFound.Status.Reason = "Namespace [ " + nsName + " ] Deleted"
		if err := r.Status().Update(context.TODO(), nscFound); err != nil {
			reqLogger.Error(err, "Failed to Update NamespaceClaim [ "+nscName+" ]")
			panic("Failed to Update NamespaceClaim [ " + nscName + " ]")
		} else {
			reqLogger.Info("Update NamespaceClaim [ " + nscName + " ] Success")
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

// postRequestToHypercloudApiServer requests POST API to Hypercloud API Server,
// which broadcasts namespace change event to all websocket clients.
// The events has 3 types; "ADDED", "DELETED", "MODIFIED".
func (r *NamespaceReconciler) postRequestToHypercloudApiServer(ns *v1.Namespace, event string) (string, error) {
	reqLogger := r.Log

	byte, err := json.Marshal(ns)
	if err != nil {
		reqLogger.Error(err, " Failed to marshal")
		return "", err
	}

	var ns_body string
	switch event {
	case "ADDED":
		ns_body = `{ "type": "ADDED", "object": `
	case "DELETED":
		ns_body = `{ "type": "DELETED", "object": `
	case "MODIFIED":
		ns_body = `{ "type": "MODIFIED", "object": `
	default:
		err := goError.New("Invalid event type")
		reqLogger.Error(err, "")
		return "", err
	}

	ns_body += string(byte) + `}`

	ns_body_bytes, err := json.Marshal(ns_body)
	if err != nil {
		reqLogger.Error(err, "Failed to marshal")
		return "", err
	}
	ns_body_io_reader := bytes.NewBuffer(ns_body_bytes)

	resp, err := http.Post(url, "application/json", ns_body_io_reader)
	if err != nil {
		reqLogger.Error(err, " Failed to broadcast namespace create event")
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Error(err, " Failed to read response body")
		return "", err
	}
	return string(body), nil
}
