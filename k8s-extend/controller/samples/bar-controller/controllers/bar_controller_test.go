package controllers

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	barv1alpha1 "github.com/ginoh/bar-controller/api/v1alpha1"
)

var _ = Describe("bar controller", func() {
	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &barv1alpha1.Bar{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		reconciler := BarReconciler{
			//Client:   k8sClient,
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Microsecond)
	})

	It("should create Deployment", func() {
		bar := newBar()
		err := k8sClient.Create(ctx, bar)
		Expect(err).NotTo(HaveOccurred())

		dep := appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample-deployment"}, &dep)
		}).Should(Succeed())
		Expect(dep.Spec.Replicas).Should(Equal(pointer.Int32Ptr(3)))
		Expect(dep.Spec.Template.Spec.Containers[0].Image).Should(Equal("envoyproxy/envoy-dev:latest"))
	})

	It("should update status", func() {
		bar := newBar()
		err := k8sClient.Create(ctx, bar)
		Expect(err).NotTo(HaveOccurred())

		updated := barv1alpha1.Bar{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample"}, &updated)
			if err != nil {
				return err
			}

			if updated.Status.AvailableReplicas != 0 {
				return errors.New("status should be updated")
			}
			return nil
		}).Should(Succeed())
	})
})

func newBar() *barv1alpha1.Bar {
	return &barv1alpha1.Bar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
		},
		Spec: barv1alpha1.BarSpec{
			DeploymentName: "sample-deployment",
			Replicas:       pointer.Int32(3),
		},
	}
}
