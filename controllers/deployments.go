package controllers

import (
	"context"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ensureDeployment reconciles Deployment child objects
func ensureDeployment(ctx context.Context, r *MailhogInstanceReconciler, cr *mailhogv1alpha1.MailhogInstance) (err error) {
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanDeployment)

	existingDeployment := &appsv1.Deployment{}
	if err = r.Get(ctx, name, existingDeployment); err != nil {
		if errors.IsNotFound(err) {
			deployment := deploymentNew(cr)
			return r.create(ctx, cr, logger, deployment, deploymentCreate)
		}
		logger.Error(err, failedGetExisting)
		return err
	}

	updatedDeployment, updateNeeded, err := deploymentUpdates(cr, existingDeployment)
	if err != nil {
		logger.Error(err, failedUpdateCheck)
		return err
	} else if updateNeeded {
		return r.update(ctx, cr, logger, updatedDeployment, deploymentUpdate)
	}

	logger.Info(stateEnsured)
	return nil
}

// deploymentNew returns a Deployment in the wanted state
func deploymentNew(cr *mailhogv1alpha1.MailhogInstance) (newDeployment *appsv1.Deployment) {
	template := podTemplate(cr)
	replicas := cr.Spec.Replicas
	meta := CreateMetaMaker(cr)

	deployment := &appsv1.Deployment{
		ObjectMeta: meta.GetMeta(),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: meta.GetLabels(),
			},
			Template: template,
		},
	}

	return deployment
}

// deploymentUpdates checks if a Deployment needs  to be updated
func deploymentUpdates(cr *mailhogv1alpha1.MailhogInstance, oldDeployment *appsv1.Deployment) (updatedDeployment *appsv1.Deployment, updateNeeded bool, err error) {
	newDeployment := deploymentNew(cr)

	updateNeeded, err = checkPatch(oldDeployment, newDeployment)
	if updateNeeded == true {
		return newDeployment, updateNeeded, err
	}
	return oldDeployment, updateNeeded, err
}
