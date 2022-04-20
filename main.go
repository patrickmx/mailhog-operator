/*
Copyright 2022.

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
	"flag"
	"os"
	"runtime/debug"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	"goimports.patrick.mx/mailhog-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	configFile string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))
	utilruntime.Must(mailhogv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

const (
	configFilePath       = "/operatorconfig/defaultconfig.yml"
	configFileFlag       = "config"
	configFileUsage      = "config file path"
	OlmDelegateNamespace = "OLM_TARGET_NAMESPACE"
	eventRecorderSource  = "mailhog-operator"

	errLoadConfig       = "unable to load config file"
	errCreateManager    = "unable to create new manager with config"
	errCreateController = "unable to create new controller"
	errAddHealthCheck   = "unable to add health check"
	errAddReadyCheck    = "unable to add ready check"
	errStartManager     = "unable to start manager"
)

func main() {
	flag.StringVar(&configFile, configFileFlag, configFilePath, configFileUsage)
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	logBuild()

	options, err := loadConfig()
	if err != nil {
		errExit(err, errLoadConfig)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		errExit(err, errCreateManager)
	}

	if err = (&controllers.MailhogInstanceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(eventRecorderSource),
	}).SetupWithManager(mgr); err != nil {
		errExit(err, errCreateController)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		errExit(err, errAddHealthCheck)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		errExit(err, errAddReadyCheck)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		errExit(err, errStartManager)
	}
}

func loadConfig() (options ctrl.Options, err error) {
	format := mailhogv1alpha1.OperatorConfig{}
	options = ctrl.Options{
		Scheme: scheme,
	}
	options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&format))
	if ns := delegateNamespacesOlm(); ns != "" {
		namespaces := strings.Split(ns, ",")
		if len(namespaces) > 1 {
			setupLog.Info("delegate namespaces from environment will override configfile value", "namespaces", ns)
			options.NewCache = cache.MultiNamespacedCacheBuilder(namespaces)
		} else {
			setupLog.Info("delegate namespace from environment will override configfile value", "namespace", ns)
			options.Namespace = ns
		}
	}
	setupLog.Info("mailhog-operator is configured", "configfile", configFile,
		"watching.namespace", options.Namespace,
		"leaderelection", options.LeaderElection, "leaderelection.namespace", options.LeaderElectionNamespace)
	return
}

func delegateNamespacesOlm() string {
	ns, found := os.LookupEnv(OlmDelegateNamespace)
	if !found {
		return ""
	}
	setupLog.Info("delegate namespace from environment will override configfile value")
	return ns
}

func errExit(err error, msg string) {
	setupLog.Error(err, msg)
	os.Exit(1)
}

func logBuild() {
	info, infoFound := debug.ReadBuildInfo()
	if infoFound {
		filteredInfo := make(map[string]string)
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				filteredInfo["revision"] = setting.Value
			case "vcs.time":
				filteredInfo["time"] = setting.Value
			case "vcs.modified":
				filteredInfo["modified"] = setting.Value
			}
		}
		if len(filteredInfo) > 0 {
			setupLog.Info("build info", "vcs", filteredInfo)
			return
		}
	}
	setupLog.Info("no build info found", "build", "null")
	return
}
