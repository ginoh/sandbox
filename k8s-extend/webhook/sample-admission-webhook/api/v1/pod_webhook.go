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

package v1

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

type PodWebhook struct{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &PodWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *PodWebhook) Default(ctx context.Context, obj runtime.Object) error {
	pod := obj.(*corev1.Pod)
	podlog.Info("default", "name", pod.GetName())

	// TODO(user): fill in your defaulting logic.
	annotations := pod.GetAnnotations()

	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["sample-admission-webhook/test-key"] = randomString(10)
	pod.SetAnnotations(annotations)

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate--v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups=core,resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &PodWebhook{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *PodWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	pod := obj.(*corev1.Pod)
	podlog.Info("validate create", "name", pod.GetName())

	// TODO(user): fill in your validation logic upon object creation.
	return r.validatePod(pod)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *PodWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	pod := newObj.(*corev1.Pod)
	podlog.Info("validate update", "name", pod.GetName())

	// TODO(user): fill in your validation logic upon object update.
	return r.validatePod(pod)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *PodWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	pod := obj.(*corev1.Pod)
	podlog.Info("validate delete", "name", pod.GetName())

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *PodWebhook) validatePod(obj *corev1.Pod) error {
	var errs field.ErrorList

	if obj.GetAnnotations() != nil {
		ignore, found := obj.Annotations["sample-admission-webhook/ignore"]
		if found && (ignore == "true") {
			return nil
		}
	}
	if err := r.validateName(obj); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: "core.v1", Kind: "Pod"}, obj.GetName(), errs)
		return err
	}
	return nil
}

func (r *PodWebhook) validateName(obj *corev1.Pod) *field.Error {

	if !strings.HasPrefix(obj.GetName(), "sample-") {
		return field.Invalid(field.NewPath("metadata", "name"), obj.GetName(),
			"name must be prefixed with sample-")
	}
	return nil
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
