package controllers

import (
	"context"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *MailhogInstanceReconciler) ensureService(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues("span", "service")

	// check if a service exists, if not create it
	existingService := &corev1.Service{}
	if err = r.Get(ctx, name, existingService); err != nil {
		if errors.IsNotFound(err) {
			// create new service
			service := r.serviceNew(cr)
			return r.create(ctx, cr, logger, service, serviceCreate)
		}
		logger.Error(err, "failed to get existing object")
		return &ReturnIndicator{
			Err: err,
		}
	}

	// check if the existing service needs an update
	updatedService, updateNeeded, err := r.serviceUpdates(cr, existingService)
	if err != nil {
		logger.Error(err, "failure checking if object update is needed")
		return &ReturnIndicator{
			Err: err,
		}
	} else if updateNeeded {
		return r.update(ctx, cr, logger, updatedService, serviceUpdate)
	}

	logger.Info("object state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) serviceNew(cr *mailhogv1alpha1.MailhogInstance) (newService *corev1.Service) {
	meta := CreateMetaMaker(cr)

	service := &corev1.Service{
		ObjectMeta: meta.GetMeta(false),
		Spec: corev1.ServiceSpec{
			Selector: meta.GetLabels(true),
			Ports: []corev1.ServicePort{
				{
					Port: portSmtp,
					Name: portSmtpName,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: portSmtp,
					},
				},
				{
					Port: portWeb,
					Name: portWebName,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: portWeb,
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

	updateNeeded, err = checkPatch(oldService, newService)
	if updateNeeded == true {
		return newService, updateNeeded, err
	}
	return oldService, updateNeeded, err
}
