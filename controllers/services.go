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
)

func (r *MailhogInstanceReconciler) ensureService(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	// check if a service exists, if not create it
	existingService := &corev1.Service{}
	if err = r.Get(ctx, name, existingService); err != nil {
		if errors.IsNotFound(err) {
			// create new service
			service := r.serviceNew(cr)
			return r.create(ctx, cr, logger, "service", service, serviceCreate)
		}
		logger.Error(err, "failed to get service")
		return &ReturnIndicator{
			Err: err,
		}
	}

	// check if the existing service needs an update
	updatedService, updateNeeded, err := r.serviceUpdates(cr, existingService)
	if err != nil {
		logger.Error(err, "failure checking if a service update is needed")
		return &ReturnIndicator{
			Err: err,
		}
	} else if updateNeeded {
		return r.update(ctx, cr, logger, "service", updatedService, serviceUpdate)
	}

	logger.Info("service state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) serviceNew(cr *mailhogv1alpha1.MailhogInstance) (newService *corev1.Service) {
	labels := labelsForCr(cr.Name)
	if cr.Spec.BackingResource == "deploymentConfig" {
		labels["deploymentconfig"] = cr.Name
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCr(cr.Name),
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port: 1025,
					Name: "smtp",
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1025,
					},
				},
				{
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

func (r *MailhogInstanceReconciler) serviceUpdates(cr *mailhogv1alpha1.MailhogInstance, oldService *corev1.Service) (updatedService *corev1.Service, updateNeeded bool, err error) {
	newService := r.serviceNew(cr)

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
		}
		return newService, true, nil

	}

	return oldService, false, nil
}
