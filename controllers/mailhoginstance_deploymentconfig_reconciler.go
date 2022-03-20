package controllers

import (
	"context"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	ocappsv1 "github.com/openshift/api/apps/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type (
	DeploymentConfigReturn struct {
		RequeueAfter time.Duration
		Err          error
	}
)

func (r *MailhogInstanceReconciler) ensureDeploymentConfig(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *DeploymentConfigReturn {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	// DeploymentConfig check
	{
		if cr.Spec.BackingResource == "deploymentConfig" {

			// check if a DC already exists, if not create it
			existingDeploymentConfig := &ocappsv1.DeploymentConfig{}
			if err = r.Get(ctx, name, existingDeploymentConfig); err != nil {
				if errors.IsNotFound(err) {
					deploymentConfig := r.deploymentConfigNew(cr)
					if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(deploymentConfig); err != nil {
						logger.Error(err, "cant annotate deploymentConfig with lastApplied state")
						return &DeploymentConfigReturn{
							Err: err,
						}
					}
					if err = ctrl.SetControllerReference(cr, deploymentConfig, r.Scheme); err != nil {
						logger.Error(err, "cant set owner reference of new deploymentConfig")
						return &DeploymentConfigReturn{
							Err: err,
						}
					}
					if err = r.Create(ctx, deploymentConfig); err != nil {
						logger.Error(err, "failed creating deploymentConfig")
						return &DeploymentConfigReturn{
							Err: err,
						}
					}
					logger.Info("created new DeploymentConfig")
					deploymentConfigCreate.Inc()
					return &DeploymentConfigReturn{
						RequeueAfter: requeueTime,
					}
				} else {
					logger.Error(err, "failed to get deploymentConfig")
					return &DeploymentConfigReturn{
						Err: err,
					}
				}
			} else {

				// check if the existing DC needs an update
				updatedDeploymentConfig, updateNeeded, err := r.deploymentConfigUpdates(cr, existingDeploymentConfig)
				if err != nil {
					logger.Error(err, "failed to check if deploymentConfig needs an update")
					return &DeploymentConfigReturn{
						Err: err,
					}
				} else if updateNeeded {
					if err = ctrl.SetControllerReference(cr, updatedDeploymentConfig, r.Scheme); err != nil {
						logger.Error(err, "cant set owner reference of updated deploymentConfig")
						return &DeploymentConfigReturn{
							Err: err,
						}
					}
					if err = r.Update(ctx, updatedDeploymentConfig); err != nil {
						logger.Error(err, "cant update deploymentConfig")
						return &DeploymentConfigReturn{
							Err: err,
						}
					}
					logger.Info("updated existing deploymentConfig")
					deploymentConfigUpdate.Inc()
					r.Recorder.Event(updatedDeploymentConfig, corev1.EventTypeNormal, "SuccessEvent", "deploymentConfig updated")
					return &DeploymentConfigReturn{
						RequeueAfter: requeueTime,
					}
				}
			}
		} else {

			toBeDeletedDeploymentConfig := &ocappsv1.DeploymentConfig{}
			if err = r.Get(ctx, name, toBeDeletedDeploymentConfig); err != nil {
				if !errors.IsNotFound(err) {
					logger.Error(err, "cant get to-be-removed deploymentConfig")
					return &DeploymentConfigReturn{
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
					return &DeploymentConfigReturn{
						Err: err,
					}
				}
				logger.Info("removed obsolete deploymentConfig")
				deploymentConfigDelete.Inc()
				return &DeploymentConfigReturn{
					RequeueAfter: requeueTime,
				}
			}
		}
	}

	logger.Info("deploymentConfig state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) deploymentConfigNew(instance *mailhogv1alpha1.MailhogInstance) (newDeployment *ocappsv1.DeploymentConfig) {
	labels := labelsForCr(instance.Name)
	labels["deploymentconfig"] = instance.Name
	env := envForCr(instance)
	ports := portsForCr()
	image := instance.Spec.Image
	replicas := instance.Spec.Replicas
	isExplicitlyFalse := false
	tenMinutes := int64(600)
	none := intstr.FromInt(0)
	two := intstr.FromInt(2)

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("200m")
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("150Mi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("200m")
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("150Mi")

	socketProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(1025),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	httpProbe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromInt(8025),
				Path:   "/api/v2/messages?limit=1",
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      2,
		PeriodSeconds:       5,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	deploymentConfig := &ocappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labelsForCr(instance.Name),
		},
		Spec: ocappsv1.DeploymentConfigSpec{
			Replicas:        replicas,
			Selector:        labels,
			MinReadySeconds: 30,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:           "mailhog",
							Image:          image,
							Ports:          ports,
							Env:            env,
							Resources:      resources,
							LivenessProbe:  socketProbe,
							StartupProbe:   socketProbe,
							ReadinessProbe: httpProbe,
						},
					},
					AutomountServiceAccountToken: &isExplicitlyFalse,
				},
			},
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

	if instance.Spec.Settings.Storage == "maildir" {
		if instance.Spec.Settings.StorageMaildir.Path != "" {
			podVolumes := make([]corev1.Volume, 0)
			containerVolMounts := make([]corev1.VolumeMount, 0)

			podVolumes = append(podVolumes, corev1.Volume{
				Name: "maildir-storage",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			})
			containerVolMounts = append(containerVolMounts, corev1.VolumeMount{
				Name:      "maildir-storage",
				MountPath: instance.Spec.Settings.StorageMaildir.Path,
			})

			deploymentConfig.Spec.Template.Spec.Volumes = podVolumes
			deploymentConfig.Spec.Template.Spec.Containers[0].VolumeMounts = containerVolMounts
		}
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
		} else {
			return newDC, true, nil
		}
	}

	return oldDC, false, nil

}
