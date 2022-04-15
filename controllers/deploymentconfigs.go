package controllers

import (
	"context"

	ocappsv1 "github.com/openshift/api/apps/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *MailhogInstanceReconciler) ensureDeploymentConfig(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanDeploymentConfig)

	if cr.Spec.BackingResource == mailhogv1alpha1.DeploymentConfigBacking {

		existingDeploymentConfig := &ocappsv1.DeploymentConfig{}
		if err = r.Get(ctx, name, existingDeploymentConfig); err != nil {
			if errors.IsNotFound(err) {
				// create new deploymentConfig
				deploymentConfig := r.deploymentConfigNew(cr)
				return r.create(ctx, cr, logger, deploymentConfig, deploymentConfigCreate)
			}
			logger.Error(err, failedGetExisting)
			return &ReturnIndicator{
				Err: err,
			}
		}

		updatedDeploymentConfig, updateNeeded, err := r.deploymentConfigUpdates(cr, existingDeploymentConfig)
		if err != nil {
			logger.Error(err, failedUpdateCheck)
			return &ReturnIndicator{
				Err: err,
			}
		} else if updateNeeded {
			return r.update(ctx, cr, logger, updatedDeploymentConfig, deploymentUpdate)
		}
	} else {

		toBeDeletedDeploymentConfig := &ocappsv1.DeploymentConfig{}
		if indicator := r.delete(ctx, name, toBeDeletedDeploymentConfig, logger, deploymentConfigDelete); indicator != nil {
			return indicator
		}
	}

	logger.Info(stateEnsured)
	return nil
}

func (r *MailhogInstanceReconciler) deploymentConfigNew(cr *mailhogv1alpha1.MailhogInstance) (newDeployment *ocappsv1.DeploymentConfig) {
	podTemplate := r.podTemplate(cr)
	meta := CreateMetaMaker(cr)
	podTemplate.Labels["deploymentconfig"] = cr.Name
	replicas := cr.Spec.Replicas
	tenMinutes := int64(600)
	none := intstr.FromInt(0)
	two := intstr.FromInt(2)

	deploymentConfig := &ocappsv1.DeploymentConfig{
		ObjectMeta: meta.GetMeta(false),
		Spec: ocappsv1.DeploymentConfigSpec{
			Replicas:        replicas,
			Selector:        meta.GetLabels(true),
			MinReadySeconds: 30,
			Template:        &podTemplate,
			Strategy: ocappsv1.DeploymentStrategy{
				Type: ocappsv1.DeploymentStrategyTypeRolling,
				RollingParams: &ocappsv1.RollingDeploymentStrategyParams{
					TimeoutSeconds: &tenMinutes,
					MaxUnavailable: &none,
					MaxSurge:       &two,
				},
			},
			Triggers: ocappsv1.DeploymentTriggerPolicies{
				ocappsv1.DeploymentTriggerPolicy{
					Type: ocappsv1.DeploymentTriggerOnConfigChange,
				},
			},
		},
	}

	return deploymentConfig
}

func (r *MailhogInstanceReconciler) deploymentConfigUpdates(cr *mailhogv1alpha1.MailhogInstance, oldDC *ocappsv1.DeploymentConfig) (updatedDeploymentConfig *ocappsv1.DeploymentConfig, updateNeeded bool, err error) {
	newDC := r.deploymentConfigNew(cr)

	updateNeeded, err = checkPatch(oldDC, newDC)
	if updateNeeded == true {
		return newDC, updateNeeded, err
	}
	return oldDC, updateNeeded, err
}
