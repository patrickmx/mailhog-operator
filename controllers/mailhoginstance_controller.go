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
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

// MailhogInstanceReconciler reconciles a MailhogInstance object
type MailhogInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	lastApplied = "mailhog.operators.patrick.mx/last-applied"
)

var (
	// default ReconcileAfter value if used
	requeueTime = time.Duration(30) * time.Second
)

//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances,verbs=*
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/status,verbs=*
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/scale,verbs=*
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/finalizers,verbs=*
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=*
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=*
//+kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs,verbs=*
//+kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs/status,verbs=*
//+kubebuilder:rbac:groups="",resources=events,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MailhogInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	ns := req.NamespacedName.Namespace
	name := req.NamespacedName.Name
	logger := log.FromContext(ctx, "ns", ns, "cr", name)

	if name == "" {
		logger.Info("empty round, stopping")
		return ctrl.Result{}, nil
	} else {
		logger.Info("starting reconcile")
	}

	// Get latest CR version
	cr := &mailhogv1alpha1.MailhogInstance{}
	if err = r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("cr not found, probably it was deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get cr")
		return ctrl.Result{}, err
	}

	// Deployment related checks
	{
		if wantsReturn := r.ensureDeployment(ctx, cr, logger); wantsReturn != nil {
			if wantsReturn.Err != nil {
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: wantsReturn.RequeueAfter}, nil
			}
		}
	}

	// DeploymentConfig related checks
	{
		if wantsReturn := r.ensureDeploymentConfig(ctx, cr, logger); wantsReturn != nil {
			if wantsReturn.Err != nil {
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: wantsReturn.RequeueAfter}, nil
			}
		}
	}

	// Service related checks
	{
		if wantsReturn := r.ensureService(ctx, cr, logger); wantsReturn != nil {
			if wantsReturn.Err != nil {
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: wantsReturn.RequeueAfter}, nil
			}
		}
	}

	// Route related checks
	{
		if wantsReturn := r.ensureRoute(ctx, cr, logger); wantsReturn != nil {
			if wantsReturn.Err != nil {
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: wantsReturn.RequeueAfter}, nil
			}
		}
	}

	// Update CR Status
	{
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(cr.Namespace),
			client.MatchingLabels(labelsForCr(cr.Name)),
		}
		if err = r.List(ctx, podList, listOpts...); err != nil {
			logger.Error(err, "Failed to list pods")
			return ctrl.Result{}, err
		}
		podNames := getPodNames(podList.Items)

		if !reflect.DeepEqual(podNames, cr.Status.Pods) {
			mailhogUpdate := &mailhogv1alpha1.MailhogInstance{}
			if err := r.Get(ctx, req.NamespacedName, mailhogUpdate); err != nil {
				logger.Error(err, "Failed to get latest cr version before update")
				return ctrl.Result{}, err
			} else {
				mailhogUpdate.Status.Pods = podNames
				mailhogUpdate.Status.PodCount = len(podNames)
				mailhogUpdate.Status.LabelSelector = textLabelsForCr(cr.Name)
				if err := r.Status().Update(ctx, mailhogUpdate); err != nil {
					logger.Error(err, "Failed to update cr status")
					return ctrl.Result{}, err
				}
				logger.Info("updated cr status")
				crUpdate.Inc()
			}
		}
	}

	return ctrl.Result{RequeueAfter: requeueTime}, nil
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
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForPod),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
