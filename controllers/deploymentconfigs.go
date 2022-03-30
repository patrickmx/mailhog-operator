package controllers

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	ocappsv1 "github.com/openshift/api/apps/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) ensureDeploymentConfig(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	if cr.Spec.BackingResource == mailhogv1alpha1.DeploymentConfigBacking {

		// check if a DC already exists, if not create it
		existingDeploymentConfig := &ocappsv1.DeploymentConfig{}
		if err = r.Get(ctx, name, existingDeploymentConfig); err != nil {
			if errors.IsNotFound(err) {
				deploymentConfig := r.deploymentConfigNew(cr)
				if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(deploymentConfig); err != nil {
					logger.Error(err, "cant annotate deploymentConfig with lastApplied state")
					return &ReturnIndicator{
						Err: err,
					}
				}
				if err = ctrl.SetControllerReference(cr, deploymentConfig, r.Scheme); err != nil {
					logger.Error(err, "cant set owner reference of new deploymentConfig")
					return &ReturnIndicator{
						Err: err,
					}
				}
				if err = r.Create(ctx, deploymentConfig); err != nil {
					logger.Error(err, "failed creating deploymentConfig")
					return &ReturnIndicator{
						Err: err,
					}
				}
				logger.Info("created new DeploymentConfig")
				deploymentConfigCreate.Inc()
				return &ReturnIndicator{}
			}
			logger.Error(err, "failed to get deploymentConfig")
			return &ReturnIndicator{
				Err: err,
			}
		}

		// check if the existing DC needs an update
		updatedDeploymentConfig, updateNeeded, err := r.deploymentConfigUpdates(cr, existingDeploymentConfig)
		if err != nil {
			logger.Error(err, "failed to check if deploymentConfig needs an update")
			return &ReturnIndicator{
				Err: err,
			}
		} else if updateNeeded {
			if err = ctrl.SetControllerReference(cr, updatedDeploymentConfig, r.Scheme); err != nil {
				logger.Error(err, "cant set owner reference of updated deploymentConfig")
				return &ReturnIndicator{
					Err: err,
				}
			}
			if err = r.Update(ctx, updatedDeploymentConfig); err != nil {
				logger.Error(err, "cant update deploymentConfig")
				return &ReturnIndicator{
					Err: err,
				}
			}
			logger.Info("updated existing deploymentConfig")
			deploymentConfigUpdate.Inc()
			r.Recorder.Event(updatedDeploymentConfig, corev1.EventTypeNormal, "SuccessEvent", "deploymentConfig updated")
			return &ReturnIndicator{}
		}
	} else {

		toBeDeletedDeploymentConfig := &ocappsv1.DeploymentConfig{}
		if err = r.Get(ctx, name, toBeDeletedDeploymentConfig); err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, "cant get to-be-removed deploymentConfig")
				return &ReturnIndicator{
					Err: err,
				}
			}
		} else {
			graceSeconds := int64(100)
			deleteOptions := client.DeleteOptions{
				GracePeriodSeconds: &graceSeconds,
			}
			if err = r.Delete(ctx, toBeDeletedDeploymentConfig, &deleteOptions); err != nil {
				logger.Error(err, "cont remove obsolete deploymentConfig")
				return &ReturnIndicator{
					Err: err,
				}
			}
			logger.Info("removed obsolete deploymentConfig")
			deploymentConfigDelete.Inc()
			return &ReturnIndicator{}
		}
	}

	logger.Info("deploymentConfig state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) deploymentConfigNew(cr *mailhogv1alpha1.MailhogInstance) (newDeployment *ocappsv1.DeploymentConfig) {
	podTemplate := r.podTemplate(cr)
	labels := labelsForCr(cr.Name)
	labels["deploymentconfig"] = cr.Name
	podTemplate.Labels["deploymentconfig"] = cr.Name
	replicas := cr.Spec.Replicas
	tenMinutes := int64(600)
	none := intstr.FromInt(0)
	two := intstr.FromInt(2)

	deploymentConfig := &ocappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCr(cr.Name),
		},
		Spec: ocappsv1.DeploymentConfigSpec{
			Replicas:        replicas,
			Selector:        labels,
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

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldDC, newDC, opts...)
	if err != nil {
		return oldDC, false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newDC); err != nil {
			return newDC, true, err
		}
		return newDC, true, nil
	}

	return oldDC, false, nil
}
