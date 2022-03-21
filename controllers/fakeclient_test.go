package controllers

import (
	"context"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestMailhogReconcileFakeclient(t *testing.T) {

	var (
		name  = "testee"
		ns    = "default"
		image = "test/test:previous"
	)

	mailhog := &mailhogv1alpha1.MailhogInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "mailhog.operators.patrick.mx/v1alpha1",
			Kind:       "MailhogInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: mailhogv1alpha1.MailhogInstanceSpec{
			Replicas: 2,
			Image:    image,
			Settings: mailhogv1alpha1.MailhogInstanceSettingsSpec{
				Hostname: "mailhogci",
				Storage:  "memory",
			},
			WebTrafficInlet: "none",
			BackingResource: "deployment",
		},
	}

	objects := []runtime.Object{
		mailhog,
	}

	s := runtime.NewScheme()
	utilruntime.Must(mailhogv1alpha1.AddToScheme(s))

	cl := fake.NewFakeClientWithScheme(s, objects...)

	r := &MailhogInstanceReconciler{Client: cl, Scheme: s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: ns,
		},
	}

	res, err := r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	if res != (reconcile.Result{}) {
		t.Error("reconcile did not return an empty Result")
	}

}
