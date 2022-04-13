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

package controllers

import (
	"context"
	"time"

	ocappsv1 "github.com/openshift/api/apps/v1"

	"github.com/go-logr/logr"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// MailhogInstanceReconciler reconciles a MailhogInstance object
type MailhogInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	logger   logr.Logger
}

// default ReconcileAfter value if used
var requeueTime = time.Duration(10) * time.Second

//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances,verbs=*
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/status,verbs=*
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/scale,verbs=*
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/finalizers,verbs=*
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=*
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=*
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=*
//+kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs,verbs=*
//+kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs/status,verbs=*
//+kubebuilder:rbac:groups="",resources=events,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MailhogInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx, "ns", req.NamespacedName.Namespace, "cr", req.NamespacedName.Name)
	r.logger.Info("starting reconcile")

	// Get latest CR version
	cr := &mailhogv1alpha1.MailhogInstance{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("cr not found, probably it was deleted")
			return ctrl.Result{}, nil
		}
		r.logger.Error(err, "failed to get cr")
		return ctrl.Result{}, err
	}

	// ensure child objects
	assurances := []func(context.Context, *mailhogv1alpha1.MailhogInstance) *ReturnIndicator{
		r.ensureDeployment,
		r.ensureDeploymentConfig,
		r.ensureService,
		r.ensureConfigMap,
		r.ensureRoute,
		r.ensureStatus,
	}
	for _, ensure := range assurances {
		if ri := ensure(ctx, cr); ri != nil {
			if ri.Err != nil {
				return ctrl.Result{}, ri.Err
			}
			return ctrl.Result{RequeueAfter: requeueTime}, nil
		}
	}

	r.logger.Info("reconciliation finished, nothing to do")
	return ctrl.Result{}, nil
}

func (r *MailhogInstanceReconciler) findObjectsForPod(watchedPod client.Object) []reconcile.Request {
	name := watchedPod.GetName()
	ns := watchedPod.GetNamespace()
	requests := make([]reconcile.Request, 0)

	pod := &corev1.Pod{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: name}, pod); err != nil {
		return []reconcile.Request{}
	}

	requests = append(requests, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: ns,
			Name:      pod.Labels["mailhog_cr"],
		},
	})

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *MailhogInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	patch.DefaultAnnotator = patch.NewAnnotator(lastApplied)

	return ctrl.NewControllerManagedBy(mgr).
		For(&mailhogv1alpha1.MailhogInstance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&ocappsv1.DeploymentConfig{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForPod),
			builder.WithPredicates(predicate.LabelChangedPredicate{}),
		).
		Complete(r)
}
