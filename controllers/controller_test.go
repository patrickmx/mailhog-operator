package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
			objects := []client.Object{
				cr,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{}))

			createdDeployment := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, nsname, createdDeployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdDeployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(image))
		})
	})

	Context("reconcile with a mailhog cr and a deployment", func() {
		It("should create a service", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.RouteTrafficInlet)
			deployment := getTestingDeployment(cr)
			objects := []client.Object{
				cr, deployment,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{}))

			createdService := &corev1.Service{}
			err = k8sClient.Get(ctx, nsname, createdService)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdService.Spec.Selector[crNameLabel]).Should(Equal(name))
		})
	})

	Context("reconcile with a mailhog cr, a deployment and a service", func() {
		It("should create a route", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.RouteTrafficInlet)
			deployment := getTestingDeployment(cr)
			service := getTestingService(cr)
			objects := []client.Object{
				cr, deployment, service,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{}))

			createdRoute := &routev1.Route{}
			err = k8sClient.Get(ctx, nsname, createdRoute)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdRoute.Spec.To.Name).To(Equal(cr.Name))
		})
	})

	Context("reconcile with a mailhog cr, when the route is deactivated but exists", func() {
		It("should delete the route", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
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
			Expect(res).Should(Equal(reconcile.Result{}))

			createdRoute := &routev1.Route{}
			err = k8sClient.Get(ctx, nsname, createdRoute)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("reconcile with a mailhog cr that needs a configmap for ui password", func() {
		It("should create the configmap", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
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
			Expect(res).Should(Equal(reconcile.Result{}))

			createdConfigMap := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, nsname, createdConfigMap)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdConfigMap.Data[settingsFilePasswordsName]).ToNot(BeEmpty())
		})
	})

	Context("reconcile with a mailhog cr that needs a configmap for smtp upstream", func() {
		It("should create the configmap correctly formatted", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
			cr.Spec.Settings.Files = &mailhogv1alpha1.MailhogFilesSpec{
				SmtpUpstreams: []mailhogv1alpha1.MailhogUpstreamSpec{
					{
						Name: "cornflower",
						Host: "blue",
					},
					{
						Name: "green",
						Host: "grass",
					},
					{
						Name: "black",
						Host: "hole",
					},
				},
			}
			expectedJson := `{"black":{"name":"black","host":"hole"},"cornflower":{"name":"cornflower","host":"blue"},"green":{"name":"green","host":"grass"}}`
			deployment := getTestingDeployment(cr)
			service := getTestingService(cr)
			objects := []client.Object{
				cr, deployment, service,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
			res, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{}))

			createdConfigMap := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, nsname, createdConfigMap)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdConfigMap.Data[settingsFileUpstreamsName]).To(Equal(expectedJson))
		})
	})

	Context("reconcile with a mailhog cr that specifies an illegal mount", func() {
		It("should return an error and refuse to proceed", func() {
			paths := []string{
				"/",
				"/usr", "/usr/local", "/usr/local/bin", "/usr/local/bin/MailHog",
				"/mailhog", "/mailhog/settings", "/mailhog/settings/files",
			}
			for _, path := range paths {
				cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
				cr.Spec.Settings.Files = &mailhogv1alpha1.MailhogFilesSpec{
					WebUsers: []mailhogv1alpha1.MailhogWebUserSpec{
						{
							Name:         "gOmega",
							PasswordHash: "bcrypt.gibberish",
						},
					},
				}
				cr.Spec.Settings.StorageMaildir.Path = path
				objects := []client.Object{
					cr,
				}
				k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

				r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
				_, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())

				updatedCr := &mailhogv1alpha1.MailhogInstance{}
				err = k8sClient.Get(ctx, nsname, updatedCr)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedCr.Status.Error).ToNot(BeEmpty())
				Expect(updatedCr.Status.Error).To(Equal(errConflictingMount.Error()))
			}
		})
	})

	Context("reconcile with a mailhog cr that specifies an illegal jim float", func() {
		It("should return an error and refuse to proceed", func() {
			cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
			cr.Spec.Settings.Jim.Invite = true
			cr.Spec.Settings.Jim.Accept = "aaa"
			objects := []client.Object{
				cr,
			}
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
			_, err := r.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())

			updatedCr := &mailhogv1alpha1.MailhogInstance{}
			err = k8sClient.Get(ctx, nsname, updatedCr)
			Expect(err).ToNot(HaveOccurred())
			Expect(updatedCr.Status.Error).ToNot(BeEmpty())
			Expect(updatedCr.Status.Error).To(Equal(errJimNonFloatFound.Error()))
		})
	})

	Context("reconcile with a mailhog cr that specifies an non-relative webroot", func() {
		It("should return an error and refuse to proceed", func() {
			paths := []string{"/first", "/deep/below/", "last/"}
			for _, path := range paths {
				cr := getTestingCr(nsname, image, mailhogv1alpha1.NoTrafficInlet)
				cr.Spec.Settings.WebPath = path
				objects := []client.Object{
					cr,
				}
				k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

				r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}
				_, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())

				updatedCr := &mailhogv1alpha1.MailhogInstance{}
				err = k8sClient.Get(ctx, nsname, updatedCr)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedCr.Status.Error).ToNot(BeEmpty())
				Expect(updatedCr.Status.Error).To(Equal(errWebPathNonRelative.Error()))
			}
		})
	})
})

func getTestingCr(nsname types.NamespacedName, image string, inlet mailhogv1alpha1.TrafficInletResource) *mailhogv1alpha1.MailhogInstance {
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
