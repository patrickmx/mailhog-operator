package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	deploymentCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_deployment_create",
			Help: "Number of times a reconcile created a deployment",
		},
	)
	deploymentUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_deployment_update",
			Help: "Number of times a reconcile updated a deployment",
		},
	)
	serviceCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_service_create",
			Help: "Number of times a reconcile created a service",
		},
	)
	serviceUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_service_update",
			Help: "Number of times a reconcile updated a service",
		},
	)
	routeCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_route_create",
			Help: "Number of times a reconcile created a route",
		},
	)
	routeUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_route_update",
			Help: "Number of times a reconcile updated a route",
		},
	)
	routeDelete = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_route_delete",
			Help: "Number of times a reconcile deleted a route",
		},
	)
	crUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_cr_update",
			Help: "Number of times a reconcile updated a cr",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(deploymentCreate, deploymentUpdate, serviceCreate, serviceUpdate, routeCreate, routeUpdate, routeDelete, crUpdate)
}
