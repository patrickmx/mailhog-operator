package controllers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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
				return r.create(ctx, cr, logger, "configmap", cm, confMapCreate)
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
			return r.update(ctx, cr, logger, "configmap", updatedCM, confMapUpdate)
		}

	} else {
		toBeDeletedCM := &corev1.ConfigMap{}
		if indicator := r.delete(ctx, name, toBeDeletedCM, "configmap", logger, confMapDelete); indicator != nil {
			return indicator
		}
	}

	logger.Info("configmap state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) configMapNew(cr *mailhogv1alpha1.MailhogInstance) (newConfigMap *corev1.ConfigMap) {
	mapdata := make(map[string]string)

	if len(cr.Spec.Settings.Files.SmtpUpstreams) > 0 {
		var serverLines []string
		for _, server := range cr.Spec.Settings.Files.SmtpUpstreams {
			text, _ := json.Marshal(server)
			serverLines = append(serverLines, "\""+server.Name+"\":"+string(text))
		}
		mapdata[settingsFileUpstreamsName] = "{" + strings.Join(serverLines, ",") + "}"
	}

	if len(cr.Spec.Settings.Files.WebUsers) > 0 {
		users := ""
		for _, credential := range cr.Spec.Settings.Files.WebUsers {
			users += credential.Name + ":" + credential.PasswordHash + "\n"
		}
		mapdata[settingsFilePasswordsName] = users
	}

	meta := CreateMetaMaker(cr)
	notImmutable := false
	configMap := &corev1.ConfigMap{
		ObjectMeta: meta.GetMeta(false),
		Immutable:  &notImmutable,
		Data:       mapdata,
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
