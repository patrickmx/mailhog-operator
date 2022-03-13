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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// MailhogInstanceReconciler reconciles a MailhogInstance object
type MailhogInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	lastApplied = "operators.patrick.mx/mailhog/last-applied"
)

//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mailhog.operators.patrick.mx,resources=mailhoginstances/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=*
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=*

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
		// check if a deployment exists, if not create it
		existingDeployment := &appsv1.Deployment{}
		if err = r.Get(ctx, req.NamespacedName, existingDeployment); err != nil {
			if errors.IsNotFound(err) {
				// create new deployment
				deployment := r.deploymentNew(cr)
				// annotate current version
				if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(deployment); err != nil {
					logger.Error(err, "failed to annotate deployment with initial state")
					return ctrl.Result{}, err
				}
				if err = r.Create(ctx, deployment); err != nil {
					logger.Error(err, "failed creating a new deployment")
					return ctrl.Result{}, err
				}
				logger.Info("created new deployment")
			} else {
				logger.Error(err, "failed to get deployment")
				return ctrl.Result{}, err
			}
		}

		// check if the existing deployment needs an update
		updatedDeployment, updateNeeded, err := r.deploymentUpdates(cr, existingDeployment)
		if err != nil {
			logger.Error(err, "failure checking if a deployment update is needed")
			return ctrl.Result{}, err
		} else if updateNeeded {
			if err = ctrl.SetControllerReference(cr, updatedDeployment, r.Scheme); err != nil {
				logger.Error(err, "cant set owner reference of updated deployment")
				return ctrl.Result{}, err
			}
			if err = r.Update(ctx, updatedDeployment); err != nil {
				logger.Error(err, "cant update deployment")
				return ctrl.Result{}, err
			}
			logger.Info("updated existing deployment")
		}

	}

	// Service related checks
	{
		// check if a service exists, if not create it
		existingService := &corev1.Service{}
		if err = r.Get(ctx, req.NamespacedName, existingService); err != nil {
			if errors.IsNotFound(err) {
				// create new service
				service := r.serviceNew(cr)
				// annotate current version
				if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(service); err != nil {
					logger.Error(err, "failed to annotate service with initial state")
					return ctrl.Result{}, err
				}
				if err = r.Create(ctx, service); err != nil {
					logger.Error(err, "failed creating a new service")
					return ctrl.Result{}, err
				}
				logger.Info("created new service")
			} else {
				logger.Error(err, "failed to get service")
				return ctrl.Result{}, err
			}
		}

		// check if the existing service needs an update
		updatedService, updateNeeded, err := r.serviceUpdates(cr, existingService)
		if err != nil {
			logger.Error(err, "failure checking if a service update is needed")
			return ctrl.Result{}, err
		} else if updateNeeded {
			if err = ctrl.SetControllerReference(cr, updatedService, r.Scheme); err != nil {
				logger.Error(err, "cant set owner reference of updated service")
				return ctrl.Result{}, err
			}
			if err = r.Update(ctx, updatedService); err != nil {
				logger.Error(err, "cant update service")
				return ctrl.Result{}, err
			}
			logger.Info("updated existing service")
		}
	}

	// Route related checks
	{
		if cr.Spec.WebTrafficInlet == "route" {

			// check if a route exists, if not create it
			existingRoute := &routev1.Route{}
			if err = r.Get(ctx, req.NamespacedName, existingRoute); err != nil {
				if errors.IsNotFound(err) {
					// create new route
					route := r.routeNew(cr)
					// annotate current version
					if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(route); err != nil {
						logger.Error(err, "failed to annotate route with initial state")
						return ctrl.Result{}, err
					}
					if err = r.Create(ctx, route); err != nil {
						logger.Error(err, "failed creating a new route")
						return ctrl.Result{}, err
					}
					logger.Info("created new route")
				} else {
					logger.Error(err, "failed to get route")
					return ctrl.Result{}, err
				}
			}

			// check if the existing route needs an update
			updatedRoute, updateNeeded, err := r.routeUpdates(cr, existingRoute)
			if err != nil {
				logger.Error(err, "failure checking if a route update is needed")
				return ctrl.Result{}, err
			} else if updateNeeded {
				if err = ctrl.SetControllerReference(cr, updatedRoute, r.Scheme); err != nil {
					logger.Error(err, "cant set owner reference of updated route")
					return ctrl.Result{}, err
				}
				if err = r.Update(ctx, updatedRoute); err != nil {
					logger.Error(err, "cant update route")
					return ctrl.Result{}, err
				}
				logger.Info("updated existing route")
			}

		} else {

			toBeDeletedRoute := &routev1.Route{}
			if err = r.Get(ctx, req.NamespacedName, toBeDeletedRoute); err != nil {
				if !errors.IsNotFound(err) {
					logger.Error(err, "cant get to-be-removed route")
					return ctrl.Result{}, err
				}
			} else {
				graceSeconds := int64(100)
				deleteOptions := client.DeleteOptions{
					GracePeriodSeconds: &graceSeconds,
				}
				if err = r.Delete(ctx, toBeDeletedRoute, &deleteOptions); err != nil {
					logger.Error(err, "cant remove obsolete route")
					return ctrl.Result{}, err
				}
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
				if err := r.Status().Update(ctx, mailhogUpdate); err != nil {
					logger.Error(err, "Failed to update cr status")
					return ctrl.Result{}, err
				}
				logger.Info("updated cr status")
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *MailhogInstanceReconciler) deploymentNew(instance *mailhogv1alpha1.MailhogInstance) (newDeployment *appsv1.Deployment) {
	labels := labelsForCr(instance.Name)
	env := envForCr(instance)
	ports := portsForCr()
	image := instance.Spec.Image
	replicas := instance.Spec.Replicas
	isExplicitlyFalse := false

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "mailhog",
							Image: image,
							Ports: ports,
							Env:   env,
						},
					},
					AutomountServiceAccountToken: &isExplicitlyFalse,
				},
			},
		},
	}

	if instance.Spec.Settings.Storage == "maildir" {
		if instance.Spec.Settings.StorageMaildir.Path != "" {
			podVolumes := make([]corev1.Volume, 0)
			containerVolMounts := make([]corev1.VolumeMount, 0)

			podVolumes = append(podVolumes, corev1.Volume{
				Name: "maildir-storage",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      "maildir-storage",
				MountPath: instance.Spec.Settings.StorageMaildir.Path,
			})

			deployment.Spec.Template.Spec.Volumes = podVolumes
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = containerVolMounts
		}
	}

	ctrl.SetControllerReference(instance, deployment, r.Scheme)

	return deployment
}

func (r *MailhogInstanceReconciler) deploymentUpdates(instance *mailhogv1alpha1.MailhogInstance, oldDeployment *appsv1.Deployment) (updatedDeployment *appsv1.Deployment, updateNeeded bool, err error) {
	newDeployment := r.deploymentNew(instance)

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldDeployment, newDeployment, opts...)

	if err != nil {
		return oldDeployment, false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newDeployment); err != nil {
			return newDeployment, true, err
		} else {
			return newDeployment, true, nil
		}
	}

	return oldDeployment, false, nil

}

func portsForCr() (p []corev1.ContainerPort) {
	return []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			ContainerPort: 8025,
			Protocol:      "TCP",
		},
		corev1.ContainerPort{
			Name:          "smtp",
			ContainerPort: 1025,
			Protocol:      "TCP",
		},
	}
}

func envForCr(crs *mailhogv1alpha1.MailhogInstance) (e []corev1.EnvVar) {
	e = append(e, corev1.EnvVar{
		Name:  "MH_SMTP_BIND_ADDR",
		Value: "0.0.0.0:1025",
	})

	e = append(e, corev1.EnvVar{
		Name:  "MH_API_BIND_ADDR",
		Value: "0.0.0.0:8025",
	})

	e = append(e, corev1.EnvVar{
		Name:  "MH_UI_BIND_ADDR",
		Value: "0.0.0.0:8025",
	})

	if crs.Spec.Settings.Storage != "" {
		e = append(e, corev1.EnvVar{
			Name:  "MH_STORAGE",
			Value: crs.Spec.Settings.Storage,
		})
	}

	if crs.Spec.Settings.Storage == "mongodb" {
		if crs.Spec.Settings.StorageMongoDb.Uri != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MONGO_URI",
				Value: crs.Spec.Settings.StorageMongoDb.Uri,
			})
		}

		if crs.Spec.Settings.StorageMongoDb.Db != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MONGO_DB",
				Value: crs.Spec.Settings.StorageMongoDb.Db,
			})
		}

		if crs.Spec.Settings.StorageMongoDb.Collection != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MONGO_COLLECTION",
				Value: crs.Spec.Settings.StorageMongoDb.Collection,
			})
		}
	}

	if crs.Spec.Settings.Storage == "maildir" {
		if crs.Spec.Settings.StorageMaildir.Path != "" {
			e = append(e, corev1.EnvVar{
				Name:  "MH_MAILDIR_PATH",
				Value: crs.Spec.Settings.StorageMaildir.Path,
			})
		}
	}

	if crs.Spec.Settings.Hostname != "" {
		e = append(e, corev1.EnvVar{
			Name:  "MH_HOSTNAME",
			Value: crs.Spec.Settings.Hostname,
		})
	}

	if crs.Spec.Settings.CorsOrigin != "" {
		e = append(e, corev1.EnvVar{
			Name:  "MH_CORS_ORIGIN",
			Value: crs.Spec.Settings.CorsOrigin,
		})
	}

	return
}

func (r *MailhogInstanceReconciler) serviceNew(instance *mailhogv1alpha1.MailhogInstance) (newService *corev1.Service) {
	labels := labelsForCr(instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: 1025,
					Name: "smtp",
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1025,
					},
				},
				corev1.ServicePort{
					Port: 8025,
					Name: "http",
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8025,
					},
				},
			},
			Type: "ClusterIP",
		},
	}

	ctrl.SetControllerReference(instance, service, r.Scheme)

	return service
}

func (r *MailhogInstanceReconciler) serviceUpdates(instance *mailhogv1alpha1.MailhogInstance, oldService *corev1.Service) (updatedService *corev1.Service, updateNeeded bool, err error) {

	newService := r.serviceNew(instance)

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldService, newService, opts...)

	if err != nil {
		return oldService, false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newService); err != nil {
			return newService, true, err
		} else {
			return newService, true, nil
		}
	}

	return oldService, false, nil

}

func (r *MailhogInstanceReconciler) routeNew(instance *mailhogv1alpha1.MailhogInstance) (newRoute *routev1.Route) {
	labels := labelsForCr(instance.Name)

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: instance.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8025,
				},
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}

	ctrl.SetControllerReference(instance, route, r.Scheme)

	return route
}

func (r *MailhogInstanceReconciler) routeUpdates(instance *mailhogv1alpha1.MailhogInstance, oldRoute *routev1.Route) (updatedRoute *routev1.Route, updateNeeded bool, err error) {
	newRoute := r.routeNew(instance)

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldRoute, newRoute, opts...)

	if err != nil {
		return oldRoute, false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newRoute); err != nil {
			return newRoute, true, err
		} else {
			return newRoute, true, nil
		}
	}

	return oldRoute, false, nil

}

func labelsForCr(name string) map[string]string {
	return map[string]string{"app": "mailhog", "mailhog_cr": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *MailhogInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	patch.DefaultAnnotator = patch.NewAnnotator(lastApplied)

	return ctrl.NewControllerManagedBy(mgr).
		For(&mailhogv1alpha1.MailhogInstance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Complete(r)
}
