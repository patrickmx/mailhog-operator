package controllers

import (
	"context"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MailhogInstanceReconciler) ensureDeployment(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues("span", "deployment")

	if cr.Spec.BackingResource == mailhogv1alpha1.DeploymentBacking {

		// check if a deployment exists, if not create it
		existingDeployment := &appsv1.Deployment{}
		if err = r.Get(ctx, name, existingDeployment); err != nil {
			if errors.IsNotFound(err) {
				// create new deployment
				deployment := r.deploymentNew(cr)
				return r.create(ctx, cr, logger, "deployment", deployment, deploymentCreate)
			}
			logger.Error(err, "failed to get deployment")
			return &ReturnIndicator{
				Err: err,
			}
		}

		// check if the existing deployment needs an update
		updatedDeployment, updateNeeded, err := r.deploymentUpdates(cr, existingDeployment)
		if err != nil {
			logger.Error(err, "failure checking if a deployment update is needed")
			return &ReturnIndicator{
				Err: err,
			}
		} else if updateNeeded {
			return r.update(ctx, cr, logger, "deployment", updatedDeployment, deploymentUpdate)
		}
	} else {

		toBeDeletedDeployment := &appsv1.Deployment{}
		if indicator := r.delete(ctx, name, toBeDeletedDeployment, "deployment", logger, deploymentDelete); indicator != nil {
			return indicator
		}
	}

	logger.Info("deployment state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) deploymentNew(cr *mailhogv1alpha1.MailhogInstance) (newDeployment *appsv1.Deployment) {
	podTemplate := r.podTemplate(cr)
	replicas := cr.Spec.Replicas
	meta := CreateMetaMaker(cr)

	deployment := &appsv1.Deployment{
		ObjectMeta: meta.GetMeta(false),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: meta.GetLabels(false),
			},
			Template: podTemplate,
		},
	}

	return deployment
}

func (r *MailhogInstanceReconciler) deploymentUpdates(cr *mailhogv1alpha1.MailhogInstance, oldDeployment *appsv1.Deployment) (updatedDeployment *appsv1.Deployment, updateNeeded bool, err error) {
	newDeployment := r.deploymentNew(cr)

	updateNeeded, err = checkPatch(oldDeployment, newDeployment)
	if updateNeeded == true {
		return newDeployment, updateNeeded, err
	}
	return oldDeployment, updateNeeded, err
}
