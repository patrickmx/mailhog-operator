package controllers

import (
	"context"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *MailhogInstanceReconciler) ensureRoute(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	if cr.Spec.WebTrafficInlet == mailhogv1alpha1.RouteTrafficInlet {

		// check if a route exists, if not create it
		existingRoute := &routev1.Route{}
		if err = r.Get(ctx, name, existingRoute); err != nil {
			if errors.IsNotFound(err) {
				// create new route
				route := r.routeNew(cr)
				return r.create(ctx, cr, logger, "route", route, routeCreate)
			}
			logger.Error(err, "failed to get route")
			return &ReturnIndicator{
				Err: err,
			}
		}

		// check if the existing route needs an update
		updatedRoute, updateNeeded, err := r.routeUpdates(cr, existingRoute)
		if err != nil {
			logger.Error(err, "failure checking if a route update is needed")
			return &ReturnIndicator{
				Err: err,
			}
		} else if updateNeeded {
			return r.update(ctx, cr, logger, "route", updatedRoute, routeUpdate)
		}

	} else {

		toBeDeletedRoute := &routev1.Route{}
		if indicator := r.delete(ctx, name, toBeDeletedRoute, "route", logger, routeDelete); indicator != nil {
			return indicator
		}
	}

	logger.Info("route state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) routeNew(cr *mailhogv1alpha1.MailhogInstance) (newRoute *routev1.Route) {
	meta := CreateMetaMaker(cr)

	route := &routev1.Route{
		ObjectMeta: meta.GetMeta(false),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8025,
				},
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}

	return route
}

func (r *MailhogInstanceReconciler) routeUpdates(cr *mailhogv1alpha1.MailhogInstance, oldRoute *routev1.Route) (updatedRoute *routev1.Route, updateNeeded bool, err error) {
	newRoute := r.routeNew(cr)

	updateNeeded, err = checkPatch(oldRoute, newRoute)
	if updateNeeded == true {
		return newRoute, updateNeeded, err
	}
	return oldRoute, updateNeeded, err
}
