package controllers

import (
	"context"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) create(ctx context.Context,
	cr *mailhogv1alpha1.MailhogInstance,
	logger logr.Logger,
	logHint string,
	obj client.Object,
	tickFunc prometheus.Counter,
) *ReturnIndicator {
	var err error

	if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(obj); err != nil {
		logger.Error(err, "failed to annotate new object with initial state", "object", logHint)
		return &ReturnIndicator{
			Err: err,
		}
	}

	if err = ctrl.SetControllerReference(cr, obj, r.Scheme); err != nil {
		logger.Error(err, "failed to set controller ref for new object", "object", logHint)
		return &ReturnIndicator{
			Err: err,
		}
	}

	if err = r.Create(ctx, obj); err != nil {
		logger.Error(err, "failed to create new object", "object", logHint)
		return &ReturnIndicator{
			Err: err,
		}
	}

	logger.Info("created new object", "object", logHint)
	tickFunc.Inc()
	return &ReturnIndicator{}
}

func (r *MailhogInstanceReconciler) delete(ctx context.Context,
	name types.NamespacedName,
	obj client.Object,
	logHint string,
	logger logr.Logger,
	tick prometheus.Counter,
) *ReturnIndicator {
	var err error

	if err = r.Get(ctx, name, obj); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "cant check for to-be-removed object", "object", logHint)
			return &ReturnIndicator{
				Err: err,
			}
		}
	} else {
		if err = r.Delete(ctx, obj, deleteOptions(100)); err != nil {
			logger.Error(err, "cant remove obsolete object", "object", logHint)
			return &ReturnIndicator{
				Err: err,
			}
		}
		logger.Info("removed obsolete object", "object", logHint)
		tick.Inc()
		return &ReturnIndicator{}
	}

	return nil
}

func deleteOptions(seconds int) *client.DeleteOptions {
	graceSeconds := int64(seconds)
	return &client.DeleteOptions{
		GracePeriodSeconds: &graceSeconds,
	}
}

func (r *MailhogInstanceReconciler) update(ctx context.Context,
	cr *mailhogv1alpha1.MailhogInstance,
	logger logr.Logger,
	logHint string,
	obj client.Object,
	tickFunc prometheus.Counter,
) *ReturnIndicator {
	var err error

	if err = ctrl.SetControllerReference(cr, obj, r.Scheme); err != nil {
		logger.Error(err, "cant set owner reference of updated object", "object", logHint)
		return &ReturnIndicator{
			Err: err,
		}
	}
	if err = r.Update(ctx, obj); err != nil {
		if errors.IsInvalid(err) {
			if deleteErr := r.Delete(ctx, obj, deleteOptions(100)); deleteErr != nil {
				logger.Error(deleteErr, "cant remove object which failed to update", "object", logHint)
				return &ReturnIndicator{
					Err: deleteErr,
				}
			}
			logger.Error(err, "deleted object because update failed", "object", logHint)
			tickFunc.Inc()
			return &ReturnIndicator{}
		}
		logger.Error(err, "cant update object", "object", logHint)
		return &ReturnIndicator{
			Err: err,
		}
	}
	logger.Info("updated existing object", "object", logHint)
	tickFunc.Inc()

	r.Recorder.Event(obj, corev1.EventTypeNormal, "SuccessEvent", "updated by mailhog management")
	return &ReturnIndicator{}
}

func checkPatch(oldO client.Object, newO client.Object) (updateNeeded bool, err error) {
	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldO, newO, opts...)
	if err != nil {
		return false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newO); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}
