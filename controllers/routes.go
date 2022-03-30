package controllers

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MailhogInstanceReconciler) ensureRoute(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance, logger logr.Logger) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}

	// Route related checks
	{
		if cr.Spec.WebTrafficInlet == mailhogv1alpha1.RouteTrafficInlet {

			// check if a route exists, if not create it
			existingRoute := &routev1.Route{}
			if err = r.Get(ctx, name, existingRoute); err != nil {
				if errors.IsNotFound(err) {
					// create new route
					route := r.routeNew(cr)
					// annotate current version
					if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(route); err != nil {
						logger.Error(err, "failed to annotate route with initial state")
						return &ReturnIndicator{
							Err: err,
						}
					}
					if err = ctrl.SetControllerReference(cr, route, r.Scheme); err != nil {
						logger.Error(err, "cant set owner reference of new route")
						return &ReturnIndicator{
							Err: err,
						}
					}
					if err = r.Create(ctx, route); err != nil {
						logger.Error(err, "failed creating a new route")
						return &ReturnIndicator{
							Err: err,
						}
					}
					logger.Info("created new route")
					routeCreate.Inc()
					return &ReturnIndicator{}
				} else {
					logger.Error(err, "failed to get route")
					return &ReturnIndicator{
						Err: err,
					}
				}
			} else {

				// check if the existing route needs an update
				updatedRoute, updateNeeded, err := r.routeUpdates(cr, existingRoute)
				if err != nil {
					logger.Error(err, "failure checking if a route update is needed")
					return &ReturnIndicator{
						Err: err,
					}
				} else if updateNeeded {
					if err = ctrl.SetControllerReference(cr, updatedRoute, r.Scheme); err != nil {
						logger.Error(err, "cant set owner reference of updated route")
						return &ReturnIndicator{
							Err: err,
						}
					}
					if err = r.Update(ctx, updatedRoute); err != nil {
						logger.Error(err, "cant update route")
						return &ReturnIndicator{
							Err: err,
						}
					}
					logger.Info("updated existing route")
					routeUpdate.Inc()
					r.Recorder.Event(updatedRoute, corev1.EventTypeNormal, "SuccessEvent", "route updated")
					return &ReturnIndicator{}
				}
			}

		} else {

			toBeDeletedRoute := &routev1.Route{}
			if err = r.Get(ctx, name, toBeDeletedRoute); err != nil {
				if !errors.IsNotFound(err) {
					logger.Error(err, "cant get to-be-removed route")
					return &ReturnIndicator{
						Err: err,
					}
				}
			} else {
				graceSeconds := int64(100)
				deleteOptions := client.DeleteOptions{
					GracePeriodSeconds: &graceSeconds,
				}
				if err = r.Delete(ctx, toBeDeletedRoute, &deleteOptions); err != nil {
					logger.Error(err, "cant remove obsolete route")
					return &ReturnIndicator{
						Err: err,
					}
				}
				logger.Info("removed obsolete route")
				routeDelete.Inc()
				return &ReturnIndicator{}
			}
		}
	}

	logger.Info("route state ensured")
	return nil
}

func (r *MailhogInstanceReconciler) routeNew(cr *mailhogv1alpha1.MailhogInstance) (newRoute *routev1.Route) {
	labels := labelsForCr(cr.Name)

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
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

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(oldRoute, newRoute, opts...)
	if err != nil {
		return oldRoute, false, err
	}

	if !patchResult.IsEmpty() {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(newRoute); err != nil {
			return newRoute, true, err
		} else {
			return newRoute, true, nil
		}
	}

	return oldRoute, false, nil
}
