/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha1

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Bar Webhook", func() {

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &Bar{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
	})
	Context("mutating", func() {
		It("should mutate a Bar", func() {
			mutateTest(filepath.Join("testdata", "mutating", "before.yaml"), filepath.Join("testdata", "mutating", "after.yaml"))
		})
	})
	Context("validating", func() {
		It("should create a valid Bar", func() {
			validateTest(filepath.Join("testdata", "validating", "valid.yaml"), true)
		})
		It("should not create invalid Bar", func() {
			validateTest(filepath.Join("testdata", "validating", "invalid-deployment-name.yaml"), false)
		})
	})
})

func mutateTest(before string, after string) {
	ctx := context.Background()

	y, err := os.ReadFile(before)
	Expect(err).NotTo(HaveOccurred())
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	beforeBar := &Bar{}
	err = d.Decode(beforeBar)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx, beforeBar)
	Expect(err).NotTo(HaveOccurred())

	actual := &Bar{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: beforeBar.GetName(), Namespace: beforeBar.GetNamespace()}, actual)
	Expect(err).NotTo(HaveOccurred())

	y, err = os.ReadFile(after)
	Expect(err).NotTo(HaveOccurred())
	d = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	afterBar := &Bar{}
	err = d.Decode(afterBar)
	Expect(err).NotTo(HaveOccurred())

	Expect(actual.Spec).Should(Equal(afterBar.Spec))
}

func validateTest(file string, valid bool) {
	ctx := context.Background()
	y, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	bar := &Bar{}
	err = d.Decode(bar)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx, bar)

	if valid {
		Expect(err).NotTo(HaveOccurred(), "Bar: %v", bar)
	} else {
		Expect(err).To(HaveOccurred(), "Bar: %v", bar)
		statusErr := &apierrors.StatusError{}
		Expect(errors.As(err, &statusErr)).To(BeTrue())
	}
}
