package controllers

import (
	"context"
	"time"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type (
	DeploymentReturn struct {
		RequeueAfter time.Duration
		Err          error
	}
)

func (r *MailhogInstanceReconciler) ensureDeployment(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *DeploymentReturn {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	// Deployment related checks
	{
		if cr.Spec.BackingResource == mailhogv1alpha1.DeploymentBacking {

			// check if a deployment exists, if not create it
			existingDeployment := &appsv1.Deployment{}
			if err = r.Get(ctx, name, existingDeployment); err != nil {
				if errors.IsNotFound(err) {
					// create new deployment
					deployment := r.deploymentNew(cr)
					// annotate current version
					if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(deployment); err != nil {
						logger.Error(err, "failed to annotate deployment with initial state")
						return &DeploymentReturn{
							Err: err,
						}
					}
					if err = ctrl.SetControllerReference(cr, deployment, r.Scheme); err != nil {
						logger.Error(err, "cant set owner reference of new deployment")
						return &DeploymentReturn{
							Err: err,
						}
					}
					if err = r.Create(ctx, deployment); err != nil {
						logger.Error(err, "failed creating a new deployment")
						return &DeploymentReturn{
							Err: err,
						}
					}
					logger.Info("created new deployment")
					deploymentCreate.Inc()
					return &DeploymentReturn{
						RequeueAfter: requeueTime,
					}
				} else {
					logger.Error(err, "failed to get deployment")
					return &DeploymentReturn{
						Err: err,
					}
				}
			} else {

				// check if the existing deployment needs an update
				updatedDeployment, updateNeeded, err := r.deploymentUpdates(cr, existingDeployment)
				if err != nil {
					logger.Error(err, "failure checking if a deployment update is needed")
					return &DeploymentReturn{
						Err: err,
					}
				} else if updateNeeded {
					if err = ctrl.SetControllerReference(cr, updatedDeployment, r.Scheme); err != nil {
						logger.Error(err, "cant set owner reference of updated deployment")
						return &DeploymentReturn{
							Err: err,
						}
					}
					if err = r.Update(ctx, updatedDeployment); err != nil {
						logger.Error(err, "cant update deployment")
						return &DeploymentReturn{
							Err: err,
						}
					}
					logger.Info("updated existing deployment")
					deploymentUpdate.Inc()
					r.Recorder.Event(updatedDeployment, corev1.EventTypeNormal, "SuccessEvent", "deployment updated")
					return &DeploymentReturn{
						RequeueAfter: requeueTime,
					}
				}
			}
		} else {

			toBeDeletedDeployment := &appsv1.Deployment{}
			if err = r.Get(ctx, name, toBeDeletedDeployment); err != nil {
				if !errors.IsNotFound(err) {
					logger.Error(err, "cant get to-be-removed deployment")
					return &DeploymentReturn{
						Err: err,
					}
				}
			} else {
				graceSeconds := int64(100)
				deleteOptions := client.DeleteOptions{
					GracePeriodSeconds: &graceSeconds,
				}
				if err = r.Delete(ctx, toBeDeletedDeployment, &deleteOptions); err != nil {
					logger.Error(err, "cant remove obsolete deployment")
					return &DeploymentReturn{
						Err: err,
					}
				}
				logger.Info("removed obsolete deployment")
				deploymentConfigDelete.Inc()
				return &DeploymentReturn{
					RequeueAfter: requeueTime,
				}
			}

		}
	}

	logger.Info("deployment state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) deploymentNew(instance *mailhogv1alpha1.MailhogInstance) (newDeployment *appsv1.Deployment) {
	labels := labelsForCr(instance.Name)
	env := envForCr(instance)
	ports := portsForCr()
	image := instance.Spec.Image
	replicas := instance.Spec.Replicas
	isExplicitlyFalse := false

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mailhog",
							Image: image,
							Ports: ports,
							Env:   env,
						},
					},
					AutomountServiceAccountToken: &isExplicitlyFalse,
				},
			},
		},
	}

	if instance.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage || instance.Spec.Settings.Files != nil {
		const (
			storageVolumeName  = "maildir-storage"
			settingsVolumeName = "settings-files"
		)
		podVolumes := make([]corev1.Volume, 0)
		containerVolMounts := make([]corev1.VolumeMount, 0)
		if instance.Spec.Settings.StorageMaildir.Path != "" && instance.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {

			podVolumes = append(podVolumes, corev1.Volume{
				Name: storageVolumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      storageVolumeName,
				MountPath: instance.Spec.Settings.StorageMaildir.Path,
			})

		}
		if instance.Spec.Settings.Files != nil {
			podVolumes = append(podVolumes, corev1.Volume{
				Name: settingsVolumeName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: instance.Name,
						},
					},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      settingsVolumeName,
				MountPath: settingsFilesMount,
			})

		}

		deployment.Spec.Template.Spec.Volumes = podVolumes
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = containerVolMounts
	}

	return deployment
}

func (r *MailhogInstanceReconciler) deploymentUpdates(instance *mailhogv1alpha1.MailhogInstance, oldDeployment *appsv1.Deployment) (updatedDeployment *appsv1.Deployment, updateNeeded bool, err error) {
	newDeployment := r.deploymentNew(instance)

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
		} else {
			return newDeployment, true, nil
		}
	}

	return oldDeployment, false, nil
}
