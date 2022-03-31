package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) createOrReturn(ctx context.Context,
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
	tick prometheus.Counter) *ReturnIndicator {
	var err error

	if err = r.Get(ctx, name, obj); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "cant check for to-be-removed object", "object", logHint)
			return &ReturnIndicator{
				Err: err,
			}
		}
	} else {
		graceSeconds := int64(100)
		deleteOptions := client.DeleteOptions{
			GracePeriodSeconds: &graceSeconds,
		}
		if err = r.Delete(ctx, obj, &deleteOptions); err != nil {
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
