package controllers

// TODO Perhaps you could add support for k8s Ingress

import (
	"context"

	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ensureRoute reconciles openshift Route child objects
func (r *MailhogInstanceReconciler) ensureRoute(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) (err error) {
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanRoute)

	if cr.Spec.WebTrafficInlet == mailhogv1alpha1.RouteTrafficInlet {

		existingRoute := &routev1.Route{}
		if err = r.Get(ctx, name, existingRoute); err != nil {
			if errors.IsNotFound(err) {
				route := r.routeNew(cr)
				return r.create(ctx, cr, logger, route, routeCreate)
			}
			logger.Error(err, failedGetExisting)
			return err
		}

		updatedRoute, updateNeeded, err := r.routeUpdates(cr, existingRoute)
		if err != nil {
			logger.Error(err, failedUpdateCheck)
			return err
		} else if updateNeeded {
			return r.update(ctx, cr, logger, updatedRoute, routeUpdate)
		}

	} else {

		toBeDeletedRoute := &routev1.Route{}
		if err = r.delete(ctx, name, toBeDeletedRoute, logger, routeDelete); err != nil {
			return err
		}
	}

	logger.Info(stateEnsured)
	return nil
}

// routeNew returns a Route in the wanted state
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

// routeUpdates checks if a Route needs  to be updated
func (r *MailhogInstanceReconciler) routeUpdates(cr *mailhogv1alpha1.MailhogInstance, oldRoute *routev1.Route) (updatedRoute *routev1.Route, updateNeeded bool, err error) {
	newRoute := r.routeNew(cr)

	updateNeeded, err = checkPatch(oldRoute, newRoute)
	if updateNeeded == true {
		return newRoute, updateNeeded, err
	}
	return oldRoute, updateNeeded, err
}
