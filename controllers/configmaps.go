package controllers

import (
	"context"
	"encoding/json"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) ensureConfigMap(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	if cr.Spec.Settings.Files != nil {
		existingCM := &corev1.ConfigMap{}
		if err = r.Get(ctx, name, existingCM); err != nil {
			if errors.IsNotFound(err) {
				// create new configmap
				cm := r.configMapNew(cr)
				return r.createOrReturn(ctx, cr, logger, "configmap", cm, confMapCreate)
			}
			logger.Error(err, "unknown error while checking for service existence")
			return &ReturnIndicator{
				Err: err,
			}
		}
		// check if update is needed
		updatedCM, updateNeeded, err := r.configMapUpdates(cr, existingCM)
		if err != nil {
			logger.Error(err, "failed check if update is needed")
			return &ReturnIndicator{
				Err: err,
			}
		} else if updateNeeded {
			if err = ctrl.SetControllerReference(cr, updatedCM, r.Scheme); err != nil {
				logger.Error(err, "error setting owner ref on new configmap")
				return &ReturnIndicator{
					Err: err,
				}
			}
			if err = r.Update(ctx, updatedCM); err != nil {
				logger.Error(err, "cant update configmap")
				return &ReturnIndicator{
					Err: err,
				}
			}
			logger.Info("updated existing configmap")
			confMapUpdate.Inc()
			r.Recorder.Event(updatedCM, corev1.EventTypeNormal, "SuccessEvent", "configmap updated")
			return &ReturnIndicator{}
		}

	} else {
		toBeDeletedCM := &corev1.ConfigMap{}
		if err = r.Get(ctx, name, toBeDeletedCM); err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, "cant check for to-be-removed configmap")
				return &ReturnIndicator{
					Err: err,
				}
			}
		} else {
			graceSeconds := int64(100)
			deleteOptions := client.DeleteOptions{
				GracePeriodSeconds: &graceSeconds,
			}
			if err = r.Delete(ctx, toBeDeletedCM, &deleteOptions); err != nil {
				logger.Error(err, "cant remove obsolete configmap")
				return &ReturnIndicator{
					Err: err,
				}
			}
			logger.Info("removed obsolete configmap")
			confMapDelete.Inc()
			return &ReturnIndicator{}
		}
	}

	logger.Info("configmap state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) configMapNew(cr *mailhogv1alpha1.MailhogInstance) (newConfigMap *corev1.ConfigMap) {
	mapdata := make(map[string]string)

	if len(cr.Spec.Settings.Files.SmtpUpstreams) > 0 {
		servers := "["
		for _, server := range cr.Spec.Settings.Files.SmtpUpstreams {
			text, _ := json.Marshal(server)
			servers += string(text)
		}
		servers += "]"
		mapdata["upstream.servers.json"] = servers
	}

	if len(cr.Spec.Settings.Files.WebUsers) > 0 {
		users := ""
		for _, credential := range cr.Spec.Settings.Files.WebUsers {
			users += credential.Name + ":" + credential.PasswordHash + "\n"
		}
		mapdata["users.list.bcrypt"] = users
	}

	notImmutable := false
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCr(cr.Name),
		},
		Immutable: &notImmutable,
		Data:      mapdata,
	}

	return configMap
}

func (r *MailhogInstanceReconciler) configMapUpdates(cr *mailhogv1alpha1.MailhogInstance, oldCM *corev1.ConfigMap) (updatedCM *corev1.ConfigMap, updateNeeded bool, err error) {
	newCM := r.configMapNew(cr)

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldCM, newCM, opts...)
	if err != nil {
		return oldCM, false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newCM); err != nil {
			return newCM, true, err
		}
		return newCM, true, nil
	}

	return oldCM, false, nil
}
