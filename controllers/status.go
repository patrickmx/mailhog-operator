package controllers

import (
	"context"
	"reflect"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) ensureStatus(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	meta := CreateMetaMaker(cr)
	logger := r.logger.WithValues(span, spanCr)

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(cr.Namespace),
		client.MatchingLabels(meta.GetLabels(false)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logger.Error(err, failedListPods)
		return &ReturnIndicator{
			Err: err,
		}
	}
	podNames := getPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, cr.Status.Pods) {
		update := &mailhogv1alpha1.MailhogInstance{}
		if err := r.Get(ctx, name, update); err != nil {
			logger.Error(err, failedCrRefresh)
			return &ReturnIndicator{
				Err: err,
			}
		}
		update.Status.Pods = podNames
		update.Status.PodCount = len(podNames)
		update.Status.LabelSelector = meta.GetSelector(true)
		if err := r.Status().Update(ctx, update); err != nil {
			logger.Error(err, failedCrUpdateStatus)
			return &ReturnIndicator{
				Err: err,
			}
		}
		logger.Info(updatedCrStatus)
		crUpdate.Inc()
		return &ReturnIndicator{}

	}

	logger.Info(noCrUpdateNeeded)
	return nil
}

func getPodNames(pods []corev1.Pod) []string {
	podNames := make([]string, 0)
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
