package controllers

import (
	"strings"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type metaMaker struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

// CreateMetaMaker returns a new ObjectMeta helper for the given CR
func CreateMetaMaker(cr *mailhogv1alpha1.MailhogInstance) *metaMaker {
	mm := &metaMaker{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}
	mm.Labels = defaultLabelsForCr(cr.Name)
	mm.Annotations = make(map[string]string)

	// https://openshift.github.io/openshift-origin-design/designs/developer/topology/#7-application-group
	if label := cr.Labels[partOfLabel]; label != "" {
		mm.Labels[partOfLabel] = label
	}

	if label := cr.Labels[appLabel]; label != "" {
		mm.Labels[appLabel] = label
	}

	if label := cr.Labels[componentLabel]; label != "" {
		mm.Labels[componentLabel] = label
	}

	if label := cr.Labels[kubeAppLabel]; label != "" {
		mm.Labels[kubeAppLabel] = label
	}

	if label := cr.Labels[runtimeLabel]; label != "" {
		mm.Labels[runtimeLabel] = label
	} else {
		mm.Labels[runtimeLabel] = runtimeDefaultValue
	}

	if label := cr.Labels[instanceLabel]; label != "" {
		mm.Labels[instanceLabel] = label
	} else {
		mm.Labels[instanceLabel] = cr.Name
	}

	// https://docs.openshift.com/container-platform/4.8/applications/odc-viewing-application-composition-using-topology-view.html#creating-a-visual-connection-between-components
	// https://www.redhat.com/en/blog/openshift-topology-view-milestone-towards-better-developer-experience
	if annotation := cr.Annotations[connectsToAnnotation]; annotation != "" {
		mm.Annotations[connectsToAnnotation] = annotation
	}

	if annotation := cr.Annotations[vcsUriAnnotation]; annotation != "" {
		mm.Annotations[vcsUriAnnotation] = annotation
	}

	return mm
}

// GetMeta returns the ObjectMeta of the CR embedded in the MetaMaker
func (mm *metaMaker) GetMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        mm.Name,
		Namespace:   mm.Namespace,
		Labels:      mm.Labels,
		Annotations: mm.Annotations,
	}
}

// GetLabels returns the map of label of the CR embedded in the MetaMaker
func (mm *metaMaker) GetLabels() map[string]string {
	return mm.GetMeta().Labels
}

// GetSelector returns the label selector of the CR embedded in the MetaMaker
func (mm *metaMaker) GetSelector() (selector string) {
	var selectors []string //nolint:prealloc

	for k, v := range mm.GetLabels() {
		selectors = append(selectors, k+"="+v)
	}
	return strings.Join(selectors, ",")
}

func defaultLabelsForCr(name string) map[string]string {
	return map[string]string{
		crTypeLabel:    crTypeValue,
		managedByLabel: operatorValue,
		createdByLabel: operatorValue,
		crNameLabel:    name,
	}
}
