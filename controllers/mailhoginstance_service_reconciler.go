package controllers

import (
	"context"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

type (
	ServiceReturn struct {
		RequeueAfter time.Duration
		Err          error
	}
)

func (r *MailhogInstanceReconciler) ensureService(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ServiceReturn {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	// Service related checks
	{
		// check if a service exists, if not create it
		existingService := &corev1.Service{}
		if err = r.Get(ctx, name, existingService); err != nil {
			if errors.IsNotFound(err) {
				// create new service
				service := r.serviceNew(cr)
				// annotate current version
				if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(service); err != nil {
					logger.Error(err, "failed to annotate service with initial state")
					return &ServiceReturn{
						Err: err,
					}
				}
				if err = ctrl.SetControllerReference(cr, service, r.Scheme); err != nil {
					logger.Error(err, "cant set owner reference of new service")
					return &ServiceReturn{
						Err: err,
					}
				}
				if err = r.Create(ctx, service); err != nil {
					logger.Error(err, "failed creating a new service")
					return &ServiceReturn{
						Err: err,
					}
				}
				logger.Info("created new service")
				serviceCreate.Inc()
				return &ServiceReturn{
					RequeueAfter: requeueTime,
				}
			} else {
				logger.Error(err, "failed to get service")
				return &ServiceReturn{
					Err: err,
				}
			}
		} else {

			// check if the existing service needs an update
			updatedService, updateNeeded, err := r.serviceUpdates(cr, existingService)
			if err != nil {
				logger.Error(err, "failure checking if a service update is needed")
				return &ServiceReturn{
					Err: err,
				}
			} else if updateNeeded {
				if err = ctrl.SetControllerReference(cr, updatedService, r.Scheme); err != nil {
					logger.Error(err, "cant set owner reference of updated service")
					return &ServiceReturn{
						Err: err,
					}
				}
				if err = r.Update(ctx, updatedService); err != nil {
					logger.Error(err, "cant update service")
					return &ServiceReturn{
						Err: err,
					}
				}
				logger.Info("updated existing service")
				serviceUpdate.Inc()
				r.Recorder.Event(updatedService, corev1.EventTypeNormal, "SuccessEvent", "service updated")
				return &ServiceReturn{
					RequeueAfter: requeueTime,
				}
			}
		}
	}

	logger.Info("service state ensured")
	return nil
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