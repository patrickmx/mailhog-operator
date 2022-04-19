package controllers

import (
	"context"
	"encoding/json"
	"strings"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// ensureConfigMap reconciles ConfigMap child objects
func (r *MailhogInstanceReconciler) ensureConfigMap(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanConfigMap)

	if cr.Spec.Settings.Files != nil {
		existingCM := &corev1.ConfigMap{}
		if err = r.Get(ctx, name, existingCM); err != nil {
			if errors.IsNotFound(err) {
				cm := r.configMapNew(cr)
				return r.create(ctx, cr, logger, cm, confMapCreate)
			}
			logger.Error(err, failedGetExisting)
			return &ReturnIndicator{
				Err: err,
			}
		}
		updatedCM, updateNeeded, err := r.configMapUpdates(cr, existingCM)
		if err != nil {
			logger.Error(err, failedUpdateCheck)
			return &ReturnIndicator{
				Err: err,
			}
		} else if updateNeeded {
			return r.update(ctx, cr, logger, updatedCM, confMapUpdate)
		}

	} else {
		toBeDeletedCM := &corev1.ConfigMap{}
		if indicator := r.delete(ctx, name, toBeDeletedCM, logger, confMapDelete); indicator != nil {
			return indicator
		}
	}

	logger.Info(stateEnsured)
	return nil
}

// configMapNew returns a ConfigMap in the wanted state
func (r *MailhogInstanceReconciler) configMapNew(cr *mailhogv1alpha1.MailhogInstance) (newConfigMap *corev1.ConfigMap) {
	data := make(map[string]string)

	if len(cr.Spec.Settings.Files.SmtpUpstreams) > 0 {
		// TODO write test & check if there i a better way to fiddle this together
		var serverLines []string
		for _, server := range cr.Spec.Settings.Files.SmtpUpstreams {
			text, _ := json.Marshal(server)
			serverLines = append(serverLines, "\""+server.Name+"\":"+string(text))
		}
		data[settingsFileUpstreamsName] = "{" + strings.Join(serverLines, ",") + "}"
	}

	if len(cr.Spec.Settings.Files.WebUsers) > 0 {
		users := ""
		for _, credential := range cr.Spec.Settings.Files.WebUsers {
			users += credential.Name + ":" + credential.PasswordHash + "\n"
		}
		data[settingsFilePasswordsName] = users
	}

	meta := CreateMetaMaker(cr)
	notImmutable := false
	configMap := &corev1.ConfigMap{
		ObjectMeta: meta.GetMeta(false),
		Immutable:  &notImmutable,
		Data:       data,
	}

	return configMap
}

// configMapUpdates checks if a ConfigMap needs  to be updated
func (r *MailhogInstanceReconciler) configMapUpdates(cr *mailhogv1alpha1.MailhogInstance, oldCM *corev1.ConfigMap) (updatedCM *corev1.ConfigMap, updateNeeded bool, err error) {
	newCM := r.configMapNew(cr)

	updateNeeded, err = checkPatch(oldCM, newCM)
	if updateNeeded == true {
		return newCM, updateNeeded, err
	}
	return oldCM, updateNeeded, err
}
