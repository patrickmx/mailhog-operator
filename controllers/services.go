package controllers

import (
	"context"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ensureService reconciles Service child objects
func ensureService(ctx context.Context, r *MailhogInstanceReconciler, cr *mailhogv1alpha1.MailhogInstance) (err error) {
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanService)

	existingService := &corev1.Service{}
	if err = r.Get(ctx, name, existingService); err != nil {
		if errors.IsNotFound(err) {
			service := serviceNew(cr)
			return r.create(ctx, cr, logger, service, serviceCreate)
		}
		logger.Error(err, failedGetExisting)
		return err
	}

	updatedService, updateNeeded, err := serviceUpdates(cr, existingService)
	if err != nil {
		logger.Error(err, failedUpdateCheck)
		return err
	} else if updateNeeded {
		return r.update(ctx, cr, logger, updatedService, serviceUpdate)
	}

	logger.Info(stateEnsured)
	return nil
}

// serviceNew returns a Service in the wanted state
func serviceNew(cr *mailhogv1alpha1.MailhogInstance) (newService *corev1.Service) {
	meta := CreateMetaMaker(cr)

	service := &corev1.Service{
		ObjectMeta: meta.GetMeta(),
		Spec: corev1.ServiceSpec{
			Selector: meta.GetLabels(),
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

// serviceUpdates checks if a Service needs  to be updated
func serviceUpdates(cr *mailhogv1alpha1.MailhogInstance, oldService *corev1.Service) (updatedService *corev1.Service, updateNeeded bool, err error) {
	newService := serviceNew(cr)

	updateNeeded, err = checkPatch(oldService, newService)
	if updateNeeded == true {
		return newService, updateNeeded, err
	}
	return oldService, updateNeeded, err
}
