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

// ensureStatus reconciles the status subresource of the MailhogInstance CR
func (r *MailhogInstanceReconciler) ensureStatus(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanCrStatus)

	ri, desiredStatus := r.desiredStatus(ctx, cr, logger)
	if ri != nil {
		return ri
	}

	if !reflect.DeepEqual(desiredStatus, cr.Status) {
		update := &mailhogv1alpha1.MailhogInstance{}
		if err := r.Get(ctx, name, update); err != nil {
			logger.Error(err, failedCrRefresh)
			return &ReturnIndicator{
				Err: err,
			}
		}
		update.Status = desiredStatus
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

// getPodNames will return the names of the given pods
func getPodNames(pods []corev1.Pod) []string {
	podNames := make([]string, 0)
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *MailhogInstanceReconciler) desiredStatus(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) (ri *ReturnIndicator, status mailhogv1alpha1.MailhogInstanceStatus) {
	meta := CreateMetaMaker(cr)

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(cr.Namespace),
		client.MatchingLabels(meta.GetLabels(false)),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		logger.Error(err, failedListPods)
		return &ReturnIndicator{
			Err: err,
		}, status
	}

	podNames := getPodNames(podList.Items)
	status.Pods = podNames
	status.PodCount = len(podNames)
	status.LabelSelector = meta.GetSelector(true)
	status.Error = ""
	return nil, status
}
