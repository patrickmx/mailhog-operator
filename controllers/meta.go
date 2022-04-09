package controllers

import (
	"strings"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type metaMaker struct {
	Name               string
	Namespace          string
	Labels             map[string]string
	Annotations        map[string]string
	IsDeploymentConfig bool
}

func CreateMetaMaker(cr *mailhogv1alpha1.MailhogInstance) *metaMaker {
	mm := &metaMaker{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}
	mm.Labels = defaultLabelsForCr(cr.Name)
	mm.Annotations = make(map[string]string)

	if cr.Spec.BackingResource == mailhogv1alpha1.DeploymentConfigBacking {
		mm.IsDeploymentConfig = true
	}

	if label := cr.Labels[partOfLabel]; label != "" {
		mm.Labels[partOfLabel] = label
	}

	if annotation := cr.Annotations[connectsToAnnotation]; annotation != "" {
		mm.Annotations[connectsToAnnotation] = annotation
	}

	if annotation := cr.Annotations[vcsUriAnnotation]; annotation != "" {
		mm.Annotations[vcsUriAnnotation] = annotation
	}

	return mm
}

func (mm *metaMaker) GetMeta(withDCLabel bool) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        mm.Name,
		Namespace:   mm.Namespace,
		Labels:      mm.Labels,
		Annotations: mm.Annotations,
	}
	if withDCLabel == true && mm.IsDeploymentConfig == true {
		meta.Labels["deploymentconfig"] = mm.Name
	}
	return meta
}

func (mm *metaMaker) GetLabels(withDCLabel bool) map[string]string {
	meta := mm.GetMeta(withDCLabel)
	return meta.Labels
}

func (mm *metaMaker) GetSelector(withDCLabel bool) (selector string) {
	var selectors []string //nolint:prealloc

	for k, v := range mm.GetLabels(withDCLabel) {
		selectors = append(selectors, k+"="+v)
	}
	return strings.Join(selectors, ",")
}

func defaultLabelsForCr(name string) map[string]string {
	return map[string]string{
		"app":                        "mailhog",
		"mailhog_cr":                 name,
		"app.openshift.io/runtime":   "golang",
		"app.kubernetes.io/name":     "mailhog",
		"app.kubernetes.io/instance": name,
	}
}
