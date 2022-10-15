package v1

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var testPodData = map[string]*corev1.Pod{
	"sample-pod": {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-pod",
			Namespace: "test",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "envoy",
					Image: "envoyproxy/envoy-dev",
				},
			},
		},
	},
	"invalid-pod": {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-pod",
			Namespace: "test",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "envoy",
					Image: "envoyproxy/envoy-dev",
				},
			},
		},
	},
	"invalid-pod-with-ignore-annotation": {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-pod",
			Namespace: "test",
			Annotations: map[string]string{
				"sample-admission-webhook/ignore": "true",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "envoy",
					Image: "envoyproxy/envoy-dev",
				},
			},
		},
	},
}

var _ = Describe("Pod Webhook", func() {
	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &corev1.Pod{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Mutating", func() {
		It("shoud add test-key annotation", func() {
			var err error
			pod := testPodData["sample-pod"]
			err = k8sClient.Create(ctx, pod)
			Expect(err).NotTo(HaveOccurred())

			actualPod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: pod.GetName(), Namespace: pod.GetNamespace()}, actualPod)
			Expect(err).NotTo(HaveOccurred())

			annotation := actualPod.GetAnnotations()
			Expect(annotation).NotTo(BeNil())

			v, found := annotation["sample-admission-webhook/test-key"]
			Expect(found).To(BeTrue())

			// value is 10 character random string
			Expect(len(v)).To(Equal(10))
		})
	})

	Context("Validating", func() {
		It("shoud not create Pod if there is no 'sample-' prfix", func() {
			var err error
			pod := testPodData["invalid-pod"]
			err = k8sClient.Create(ctx, pod)
			statusError := &apierrors.StatusError{}
			Expect(errors.As(err, &statusError)).To(BeTrue())
		})

		It("shoud create Pod without 'sample-' prfix if there is ignore validation", func() {
			var err error
			pod := testPodData["invalid-pod-with-ignore-annotation"]

			err = k8sClient.Create(ctx, pod)
			Expect(err).NotTo(HaveOccurred())

			actualPod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: pod.GetName(), Namespace: pod.GetNamespace()}, actualPod)
			Expect(err).NotTo(HaveOccurred())

			annotation := actualPod.GetAnnotations()
			Expect(annotation).NotTo(BeNil())

			v, found := annotation["sample-admission-webhook/test-key"]
			Expect(found).To(BeTrue())

			// value is 10 character random string
			Expect(len(v)).To(Equal(10))
		})
	})
})
