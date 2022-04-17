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

// getPodStates will return the names of the given pods
//
//nolint:gocritic
func getPodStates(pods []corev1.Pod) (states mailhogv1alpha1.PodStatus) {
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodPending {
			states.Pending = append(states.Pending, pod.Name)
		} else if pod.Status.Phase == corev1.PodFailed {
			states.Failed = append(states.Failed, pod.Name)
		} else if pod.Status.ContainerStatuses[0].RestartCount > 3 {
			states.Restarting = append(states.Restarting, pod.Name)
		} else if pod.Status.ContainerStatuses[0].Ready {
			states.Ready = append(states.Ready, pod.Name)
		} else {
			states.Other = append(states.Other, pod.Name)
		}
	}
	return
}

// getReadyPods will return the amount of ready pods
func getReadyPods(pods []corev1.Pod) int {
	ready := 0
	for _, pod := range pods {
		if pod.Status.ContainerStatuses[0].Ready {
			ready++
		}
	}
	return ready
}

// desiredStatus is sued to check the CR status subresource against the desired state
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
	status.Pods = getPodStates(podList.Items)
	status.PodCount = len(podNames)
	status.PodCount = getReadyPods(podList.Items)
	status.LabelSelector = meta.GetSelector(true)
	status.Error = ""
	return nil, status
}
