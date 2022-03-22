package controllers

import (
	"context"
	ocappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	objects := []client.Object{
		mailhog,
	}

	s := runtime.NewScheme()
	utilruntime.Must(mailhogv1alpha1.AddToScheme(s))
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(routev1.AddToScheme(s))
	utilruntime.Must(ocappsv1.AddToScheme(s))

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(objects...).Build()

	r := &MailhogInstanceReconciler{Client: cl, Scheme: s}

	nsname := types.NamespacedName{
		Name:      name,
		Namespace: ns,
	}

	req := reconcile.Request{
		NamespacedName: nsname,
	}

	res, err := r.Reconcile(context.TODO(), req)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if res != (reconcile.Result{RequeueAfter: requeueTime}) {
		t.Error("reconcile did not return an empty Result")
	}

	createdDeployment := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), nsname, createdDeployment)
	if err != nil {
		t.Fatalf("no deployment after reconcile: %v", err)
	}
	if createdDeployment.Spec.Template.Spec.Containers[0].Image != image {
		t.Fatalf("new deployment has wrong image after reconcile: %v", err)
	}

}
