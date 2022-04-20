package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	deploymentCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_deployment_create_total",
			Help: "Number of times a reconcile created a deployment",
		},
	)
	deploymentUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_deployment_update_total",
			Help: "Number of times a reconcile updated a deployment",
		},
	)
	deploymentDelete = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_deployment_delete_total",
			Help: "Number of times a reconcile deleted a deployment",
		},
	)
	serviceCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_service_create_total",
			Help: "Number of times a reconcile created a service",
		},
	)
	serviceUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_service_update_total",
			Help: "Number of times a reconcile updated a service",
		},
	)
	routeCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_route_create_total",
			Help: "Number of times a reconcile created a route",
		},
	)
	routeUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_route_update_total",
			Help: "Number of times a reconcile updated a route",
		},
	)
	routeDelete = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_route_delete_total",
			Help: "Number of times a reconcile deleted a route",
		},
	)
	crUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_cr_update_total",
			Help: "Number of times a reconcile updated a cr",
		},
	)
	confMapCreate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_configmap_create_total",
			Help: "Number of times a reconcile created a ConfigMap",
		},
	)
	confMapUpdate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_configmap_update_total",
			Help: "Number of times a reconcile updated a ConfigMap",
		},
	)
	confMapDelete = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mailhog_configmap_delete_total",
			Help: "Number of times a reconcile deleted a ConfigMap",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(deploymentCreate, deploymentUpdate, deploymentDelete)
	metrics.Registry.MustRegister(serviceCreate, serviceUpdate)
	metrics.Registry.MustRegister(routeCreate, routeUpdate, routeDelete)
	metrics.Registry.MustRegister(crUpdate)
	metrics.Registry.MustRegister(confMapCreate, confMapUpdate, confMapDelete)
}
