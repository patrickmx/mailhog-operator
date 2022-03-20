package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

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
			ctx := context.Background()
			mailhog := &v1alpha1.MailhogInstance{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mailhog.operators.patrick.mx/v1alpha1",
					Kind:       "MailhogInstance",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
				Spec: v1alpha1.MailhogInstanceSpec{
					Replicas: 2,
					Image:    image,
					Settings: v1alpha1.MailhogInstanceSettingsSpec{
						Hostname: "mailhogci",
						Storage:  "memory",
					},
					WebTrafficInlet: "none",
					BackingResource: "deployment",
				},
			}
			Expect(k8sClient.Create(ctx, mailhog)).Should(Succeed())

			lookup := types.NamespacedName{Name: name, Namespace: ns}
			createdInstance := &v1alpha1.MailhogInstance{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookup, createdInstance)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			Expect(createdInstance.Spec.Image).Should(Equal(image))

		})
	})

})
