package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ocappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
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

var _ = Describe("MailhogInstance controller", func() {
	const (
		name  = "tester"
		ns    = "default"
		image = "test/test:latest"
	)
	nsname := types.NamespacedName{
		Name:      name,
		Namespace: ns,
	}
	req := reconcile.Request{
		NamespacedName: nsname,
	}

	Context("reconcile with a mailhog cr", func() {
		It("should create a deployment", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			objects := []client.Object{
				cr,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdDeployment := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, nsname, createdDeployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
		})
	})

	Context("reconcile with a mailhog cr which uses a deploymentconfig", func() {
		It("should create a deploymentconfig", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentConfigBacking)
			objects := []client.Object{
				cr,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdDeploymentConfig := &ocappsv1.DeploymentConfig{}
			err = k8sClient.Get(ctx, nsname, createdDeploymentConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdDeploymentConfig.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
		})
	})

	Context("reconcile with a mailhog cr and a deployment", func() {
		It("should create a service", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.RouteTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			deployment := getTestingDeployment(cr)
			objects := []client.Object{
				cr, deployment,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdService := &corev1.Service{}
			err = k8sClient.Get(ctx, nsname, createdService)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdService.Spec.Selector[crNameLabel]).Should(Equal(name))
		})
	})

	Context("reconcile with a mailhog cr, a deployment and a service", func() {
		It("should create a route", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.RouteTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			deployment := getTestingDeployment(cr)
			service := getTestingService(cr)
			objects := []client.Object{
				cr, deployment, service,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdRoute := &routev1.Route{}
			err = k8sClient.Get(ctx, nsname, createdRoute)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdRoute.Spec.To.Name).To(Equal(cr.Name))
		})
	})

	Context("reconcile with a mailhog cr, when the route is deactivated but exists", func() {
		It("should delete the route", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			deployment := getTestingDeployment(cr)
			service := getTestingService(cr)
			route := getTestingRoute(cr)
			objects := []client.Object{
				cr, deployment, service, route,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdRoute := &routev1.Route{}
			err = k8sClient.Get(ctx, nsname, createdRoute)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("reconcile with a mailhog cr that needs a configmap", func() {
		It("should create the configmap", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet, mailhogv1alpha1.DeploymentBacking)
			cr.Spec.Settings.Files = &mailhogv1alpha1.MailhogFilesSpec{
				WebUsers: []mailhogv1alpha1.MailhogWebUserSpec{
					{
						Name:         "gOmega",
						PasswordHash: "bcrypt.gibberish",
					},
				},
			}
			deployment := getTestingDeployment(cr)
			service := getTestingService(cr)
			objects := []client.Object{
				cr, deployment, service,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: requeueTime}))

			createdConfigMap := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, nsname, createdConfigMap)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdConfigMap.Data[settingsFilePasswordsName]).ToNot(BeEmpty())
		})
	})
})

func getTestingCr(nsname types.NamespacedName, image string, inlet mailhogv1alpha1.TrafficInletResource, deployment mailhogv1alpha1.BackingResource) *mailhogv1alpha1.MailhogInstance {
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
				Hostname: "tester",
				Storage:  mailhogv1alpha1.MemoryStorage,
			},
			WebTrafficInlet: inlet,
			BackingResource: deployment,
		},
	}
}

func getTestingDeployment(cr *mailhogv1alpha1.MailhogInstance) *appsv1.Deployment {
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
	r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
	return r.deploymentNew(cr)
}

func getTestingService(cr *mailhogv1alpha1.MailhogInstance) *corev1.Service {
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
	r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
	return r.serviceNew(cr)
}

func getTestingRoute(cr *mailhogv1alpha1.MailhogInstance) *routev1.Route {
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
	r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
	return r.routeNew(cr)
}
