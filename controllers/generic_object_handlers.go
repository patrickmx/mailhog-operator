package controllers

import (
	"context"

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
