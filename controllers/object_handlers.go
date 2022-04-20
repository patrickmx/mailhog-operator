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

// create tries to create the given object
func (r *MailhogInstanceReconciler) create(ctx context.Context,
	cr *mailhogv1alpha1.MailhogInstance,
	logger logr.Logger,
	obj client.Object,
	tickFunc prometheus.Counter,
) *ReturnIndicator {
	var err error

	if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(obj); err != nil {
		logger.Error(err, messageFailedGetInitialObject)
		return &ReturnIndicator{
			Err: err,
		}
	}

	if err = ctrl.SetControllerReference(cr, obj, r.Scheme); err != nil {
		logger.Error(err, messageFailedSetOwnerRef)
		return &ReturnIndicator{
			Err: err,
		}
	}

	if err = r.Create(ctx, obj); err != nil {
		logger.Error(err, messageFailedCreate)
		return &ReturnIndicator{
			Err: err,
		}
	}

	logger.Info(messageCreatedObject)
	tickFunc.Inc()
	return &ReturnIndicator{}
}

// delete tries to delete the given object
func (r *MailhogInstanceReconciler) delete(ctx context.Context,
	name types.NamespacedName,
	obj client.Object,
	logger logr.Logger,
	tick prometheus.Counter,
) *ReturnIndicator {
	var err error

	if err = r.Get(ctx, name, obj); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, messageFailedGetDeletingObject)
			return &ReturnIndicator{
				Err: err,
			}
		}
	} else {
		if err = r.Delete(ctx, obj, deleteOptions(100)); err != nil {
			logger.Error(err, messageFailedDelete)
			return &ReturnIndicator{
				Err: err,
			}
		}
		logger.Info(messageDeletedObject)
		tick.Inc()
		return &ReturnIndicator{}
	}

	return nil
}

// deleteOptions returns new delete options with the given grace period
func deleteOptions(seconds int) *client.DeleteOptions {
	graceSeconds := int64(seconds)
	return &client.DeleteOptions{
		GracePeriodSeconds: &graceSeconds,
	}
}

// update tries to update the given object
func (r *MailhogInstanceReconciler) update(ctx context.Context,
	cr *mailhogv1alpha1.MailhogInstance,
	logger logr.Logger,
	obj client.Object,
	tickFunc prometheus.Counter,
) *ReturnIndicator {
	var err error

	if err = ctrl.SetControllerReference(cr, obj, r.Scheme); err != nil {
		logger.Error(err, messageFailedSetOwnerRefUpdate)
		return &ReturnIndicator{
			Err: err,
		}
	}
	if err = r.Update(ctx, obj); err != nil {
		// TODO restrict to deployment where this was actually the problem
		if errors.IsInvalid(err) {
			if deleteErr := r.Delete(ctx, obj, deleteOptions(100)); deleteErr != nil {
				logger.Error(deleteErr, messageFailedDeleteAfterInvalid)
				return &ReturnIndicator{
					Err: deleteErr,
				}
			}
			logger.Error(err, messageDeletedObjectAfterInvalid)
			tickFunc.Inc()
			return &ReturnIndicator{}
		}
		logger.Error(err, messageFailedUpdate)
		return &ReturnIndicator{
			Err: err,
		}
	}
	logger.Info(messageUpdated)
	tickFunc.Inc()

	r.Recorder.Event(cr, corev1.EventTypeNormal, "SuccessEvent", eventUpdated)
	return &ReturnIndicator{}
}

// checkPatch compares an object to its reference state
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

type (
	// ReturnIndicator is used to indicate to the main Reconcile if a Return / Requeue is needed or not
	// TODO why not just return an error or nil?
	ReturnIndicator struct {
		Err error
	}
)
