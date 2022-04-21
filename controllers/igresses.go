package controllers

import (
	"context"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func ensureIngress(ctx context.Context, r *MailhogInstanceReconciler, cr *mailhogv1alpha1.MailhogInstance) (err error) {
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanIgress)

	if cr.Spec.WebTrafficInlet == mailhogv1alpha1.IngressTrafficInlet {

		existingIngress := &networkingv1.Ingress{}
		if err = r.Get(ctx, name, existingIngress); err != nil {
			if errors.IsNotFound(err) {
				ingress := ingressNew(cr)
				return r.create(ctx, cr, logger, ingress, ingressCreate)
			}
			logger.Error(err, failedGetExisting)
			return err
		}

		updatedIngress, updateNeeded, err := ingressUpdates(cr, existingIngress)
		if err != nil {
			logger.Error(err, failedUpdateCheck)
			return err
		} else if updateNeeded {
			return r.update(ctx, cr, logger, updatedIngress, ingressUpdate)
		}

	} else {

		toBeDeletedIngress := &networkingv1.Ingress{}
		if err = r.delete(ctx, cr, name, toBeDeletedIngress, logger, ingressDelete); err != nil {
			return err
		}
	}

	logger.Info(stateEnsured)
	return nil
}

func ingressNew(cr *mailhogv1alpha1.MailhogInstance) (newIngress *networkingv1.Ingress) {
	meta := CreateMetaMaker(cr)
	prefix := networkingv1.PathTypePrefix
	rules := networkingv1.HTTPIngressRuleValue{
		Paths: []networkingv1.HTTPIngressPath{
			{
				Path:     "/" + cr.Spec.Settings.WebPath,
				PathType: &prefix,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: meta.Name,
						Port: networkingv1.ServiceBackendPort{
							Name:   portWebName,
							Number: portWeb,
						},
					},
				},
			},
		},
	}
	ingress := &networkingv1.Ingress{
		ObjectMeta: meta.GetMeta(),
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: cr.Spec.Settings.Ingress.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &rules,
					},
				},
			},
		},
	}
	if class := cr.Spec.Settings.Ingress.Class; class != "" {
		ingress.Spec.IngressClassName = &class
	}
	if secretName := cr.Spec.Settings.Ingress.TlsSecret; secretName != "" {
		secret := networkingv1.IngressTLS{
			SecretName: secretName,
			Hosts:      []string{cr.Spec.Settings.Ingress.Host},
		}
		ingress.Spec.TLS = append(ingress.Spec.TLS, secret)
	}
	return ingress
}

func ingressUpdates(cr *mailhogv1alpha1.MailhogInstance, oldIngress *networkingv1.Ingress) (updatedIngress *networkingv1.Ingress, updateNeeded bool, err error) {
	newIngress := ingressNew(cr)

	updateNeeded, err = checkPatch(oldIngress, newIngress)
	if updateNeeded == true {
		return newIngress, updateNeeded, err
	}
	return oldIngress, updateNeeded, err
}
