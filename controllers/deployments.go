package controllers

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MailhogInstanceReconciler) ensureDeployment(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

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
	labels := labelsForCr(cr.Name)
	replicas := cr.Spec.Replicas

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: podTemplate,
		},
	}

	return deployment
}

func (r *MailhogInstanceReconciler) deploymentUpdates(cr *mailhogv1alpha1.MailhogInstance, oldDeployment *appsv1.Deployment) (updatedDeployment *appsv1.Deployment, updateNeeded bool, err error) {
	newDeployment := r.deploymentNew(cr)

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
		}
		return newDeployment, true, nil
	}

	return oldDeployment, false, nil
}
