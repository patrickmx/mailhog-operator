package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ocappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
	scheme    *runtime.Scheme
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
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

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a mailhog cr", func() {
		It("should create the resources", func() {
			By("creating a new mailhog cr")
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
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

			r := &MailhogInstanceReconciler{Client: k8sClient, Scheme: scheme}

			nsname := types.NamespacedName{
				Name:      name,
				Namespace: ns,
			}
			req := reconcile.Request{
				NamespacedName: nsname,
			}
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

})
