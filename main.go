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

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/robfig/cron"

	claimv1alpha1 "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
	"github.com/tmax-cloud/hypercloud-single-operator/controllers"
	k8scontroller "github.com/tmax-cloud/hypercloud-single-operator/controllers/k8s"
	"github.com/tmax-cloud/hypercloud-single-operator/util"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(claimv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	// For Log file
	file, err := os.OpenFile(
		"/logs/operator.log",
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		os.FileMode(0644),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	w := io.MultiWriter(file, os.Stdout)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(w)))
	util.UpdateResourceList(setupLog)

	// Logging Cron Job
	cronJob := cron.New()
	cronJob.AddFunc("1 0 0 * * ?", func() {
		input, err := ioutil.ReadFile("/logs/operator.log")
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ioutil.WriteFile("/logs/operator"+time.Now().AddDate(0, 0, -1).Format("2006-01-02")+".log", input, 0644)
		if err != nil {
			fmt.Println("Error creating", "/logs/operator")
			fmt.Println(err)
			return
		}
		setupLog.Info("Log BackUp Success")
		os.Truncate("/logs/operator.log", 0)
		file.Seek(0, os.SEEK_SET)
	})
	cronJob.Start()

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "9722026a.tmax.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&k8scontroller.NamespaceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Namespace"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	if err = (&controllers.NamespaceClaimReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("NamespaceClaim"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NamespaceClaim")
		os.Exit(1)
	}
	if err = (&controllers.RoleBindingClaimReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("RoleBindingClaim"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RoleBindingClaim")
		os.Exit(1)
	}
	if err = (&controllers.ResourceQuotaClaimReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ResourceQuotaClaim"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ResourceQuotaClaim")
		os.Exit(1)
	}
	if err = (&claimv1alpha1.ResourceQuotaClaim{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ResourceQuotaClaim")
		os.Exit(1)
	}
	if err = (&claimv1alpha1.NamespaceClaim{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "NamespaceClaim")
		os.Exit(1)
	}
	if err = (&k8scontroller.ResourceQuotaReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ResourceQuota"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ResourceQuota")
		os.Exit(1)
	}
	if err = (&k8scontroller.RoleBindingReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("RoleBinding"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RoleBinding")
		os.Exit(1)
	}
	if err = (&claimv1alpha1.RoleBindingClaim{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "RoleBindingClaim")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	// Set Trial Timer
	setupLog.Info("[Main] Start Trial Namespace Timer")
	go func() {
		for {
			nsList := &v1.NamespaceList{}
			if err := mgr.GetClient().List(context.TODO(), nsList); err != nil {
				setupLog.Info("Not Ready to List Namespace Yet, Wait 1 seconds & try again")
				time.Sleep(1 * time.Second)
				continue
			}

			for _, ns := range nsList.Items {
				if ns.Labels != nil && ns.Labels["trial"] != "" && ns.Labels["period"] != "" && ns.Annotations != nil && ns.Annotations["owner"] != "" {
					setupLog.Info("[Main] Trial NameSpace : " + ns.Name)
					util.SetTrialNSTimer(&ns, mgr.GetClient(), ctrl.Log)
				}
			}
			setupLog.Info("[Main] Start Trial Namespace Timer Complete")
			break
		}
	}()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
