package controllers

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ocappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	k8sClient client.Client
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
	scheme    *runtime.Scheme
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())
	scheme = runtime.NewScheme()

	err = mailhogv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = routev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = ocappsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	cancel()
})

var _ = Describe("CronJob controller", func() {
	const (
		name  = "testee"
		ns    = "default"
		image = "test/test:previous"

		timeout  = time.Second * 3
		interval = time.Millisecond * 250
	)
	nsname := types.NamespacedName{
		Name:      name,
		Namespace: ns,
	}
	req := reconcile.Request{
		NamespacedName: nsname,
	}

	Context("Reconciling with a mailhog cr", func() {
		It("should create a deployment", func() {
			mailhog := mailhogTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			objects := []client.Object{
				mailhog,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdDeployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nsname, createdDeployment)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
		})
	})

	Context("Reconciling with a mailhog cr which uses a deploymentconfig", func() {
		It("should create a deploymentconfig", func() {
			mailhog := mailhogTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentConfigBacking)
			objects := []client.Object{
				mailhog,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdDeploymentConfig := &ocappsv1.DeploymentConfig{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nsname, createdDeploymentConfig)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdDeploymentConfig.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
		})
	})

	Context("Reconciling with a mailhog cr and a deployment", func() {
		It("should create a service", func() {
			mailhog := mailhogTestingCr(nsname, image, mailhogv1alpha1.RouteTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			deployment := mailhogTestingDeployment(mailhog)
			objects := []client.Object{
				mailhog, deployment,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdService := &corev1.Service{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nsname, createdService)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdService.Spec.Selector["mailhog_cr"]).Should(Equal(name))
		})
	})

	Context("Reconciling with a mailhog cr, a deployment and a service", func() {
		It("should create a route", func() {
			mailhog := mailhogTestingCr(nsname, image, mailhogv1alpha1.RouteTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			deployment := mailhogTestingDeployment(mailhog)
			service := mailhogTestingService(mailhog)
			objects := []client.Object{
				mailhog, deployment, service,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdRoute := &routev1.Route{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nsname, createdRoute)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdRoute.Spec.To.Name).Should(Equal(name))
		})
	})

	Context("Reconciling with a mailhog cr, when the route got removed", func() {
		It("should delete the route", func() {
			mailhog := mailhogTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			deployment := mailhogTestingDeployment(mailhog)
			service := mailhogTestingService(mailhog)
			route := mailhogTestingRoute(mailhog)
			objects := []client.Object{
				mailhog, deployment, service, route,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdRoute := &routev1.Route{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nsname, createdRoute)
				if err != nil {
					if errors.IsNotFound(err) {
						return true
					}
					return false
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func mailhogTestingCr(nsname types.NamespacedName, image string, inlet mailhogv1alpha1.TrafficInletResource, deployment mailhogv1alpha1.BackingResource) *mailhogv1alpha1.MailhogInstance {
	return &mailhogv1alpha1.MailhogInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "mailhog.operators.patrick.mx/v1alpha1",
			Kind:       "MailhogInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsname.Name,
			Namespace: nsname.Namespace,
		},
		Spec: mailhogv1alpha1.MailhogInstanceSpec{
			Replicas: 2,
			Image:    image,
			Settings: mailhogv1alpha1.MailhogInstanceSettingsSpec{
				Hostname: "mailhogci",
				Storage:  "memory",
			},
			WebTrafficInlet: inlet,
			BackingResource: deployment,
		},
	}
}

func mailhogTestingDeployment(cr *mailhogv1alpha1.MailhogInstance) *appsv1.Deployment {
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

	return r.deploymentNew(cr)
}

func mailhogTestingService(cr *mailhogv1alpha1.MailhogInstance) *corev1.Service {
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

	return r.serviceNew(cr)
}

func mailhogTestingRoute(cr *mailhogv1alpha1.MailhogInstance) *routev1.Route {
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

	return r.routeNew(cr)
}
