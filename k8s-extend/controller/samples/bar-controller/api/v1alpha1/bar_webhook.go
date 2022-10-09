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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var barlog = logf.Log.WithName("bar-resource")

func (r *Bar) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-samplecontroller-ginoh-github-io-v1alpha1-bar,mutating=true,failurePolicy=fail,sideEffects=None,groups=samplecontroller.ginoh.github.io,resources=bars,verbs=create;update,versions=v1alpha1,name=mbar.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Bar{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Bar) Default() {
	barlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.Replicas == nil {
		r.Spec.Replicas = pointer.Int32(1)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-samplecontroller-ginoh-github-io-v1alpha1-bar,mutating=false,failurePolicy=fail,sideEffects=None,groups=samplecontroller.ginoh.github.io,resources=bars,verbs=create;update,versions=v1alpha1,name=vbar.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Bar{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Bar) ValidateCreate() error {
	barlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validateBar()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Bar) ValidateUpdate(old runtime.Object) error {
	barlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.validateBar()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Bar) ValidateDelete() error {
	barlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Bar) validateDeploymentName() *field.Error {
	if len(r.Spec.DeploymentName) > 253 {
		return field.Invalid(field.NewPath("spec").
			Child("deploymentName"),
			r.Spec.DeploymentName,
			"must be no more than 253 characters")
	}
	return nil
}

func (r *Bar) validateBar() error {
	var allErrs field.ErrorList

	if err := r.validateDeploymentName(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{
		Group: "samplecontroller.k8s.io",
		Kind:  "Bar",
	}, r.Name, allErrs)
}
