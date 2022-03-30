package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) ensureStatus(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(cr.Namespace),
		client.MatchingLabels(labelsForCr(cr.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logger.Error(err, "Failed to list pods")
		return &ReturnIndicator{
			Err: err,
		}
	}
	podNames := getPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, cr.Status.Pods) {
		mailhogUpdate := &mailhogv1alpha1.MailhogInstance{}
		if err := r.Get(ctx, name, mailhogUpdate); err != nil {
			logger.Error(err, "Failed to get latest cr version before update")
			return &ReturnIndicator{
				Err: err,
			}
		}
		mailhogUpdate.Status.Pods = podNames
		mailhogUpdate.Status.PodCount = len(podNames)
		mailhogUpdate.Status.LabelSelector = textLabelsForCr(cr.Name)
		if err := r.Status().Update(ctx, mailhogUpdate); err != nil {
			logger.Error(err, "Failed to update cr status")
			return &ReturnIndicator{
				Err: err,
			}
		}
		logger.Info("updated cr status")
		crUpdate.Inc()
		return &ReturnIndicator{}

	}

	logger.Info("no cr status update required")
	return nil
}
