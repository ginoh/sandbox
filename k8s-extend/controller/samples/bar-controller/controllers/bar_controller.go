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

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	samplecontrollerv1alpha1 "github.com/ginoh/bar-controller/api/v1alpha1"
)

// BarReconciler reconciles a Bar object
type BarReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=samplecontroller.ginoh.github.io,resources=bars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=samplecontroller.ginoh.github.io,resources=bars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=samplecontroller.ginoh.github.io,resources=bars/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bar object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *BarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var bar samplecontrollerv1alpha1.Bar
	if err := r.Get(ctx, req.NamespacedName, &bar); err != nil {
		logger.Error(err, "unable to fetch Bar", "name", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// cleanup
	if err := r.cleanupOwnerResources(ctx, &bar); err != nil {
		logger.Error(err, "failed to clean up old Deployment resources for this Bar")
		return ctrl.Result{}, err
	}

	deploymentName := bar.Spec.DeploymentName
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: req.Namespace,
		},
	}

	// create
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		replicas := pointer.Int32Ptr(1)
		if bar.Spec.Replicas != nil {
			replicas = bar.Spec.Replicas
		}
		deploy.Spec.Replicas = replicas

		labels := map[string]string{
			"app":        "envoy",
			"controller": req.Name,
		}

		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		if deploy.Spec.Template.ObjectMeta.Labels == nil {
			deploy.Spec.Template.ObjectMeta.Labels = labels
		}

		containers := []corev1.Container{
			{
				Name:  "envoy",
				Image: "envoyproxy/envoy-dev:latest",
			},
		}

		if deploy.Spec.Template.Spec.Containers == nil {
			deploy.Spec.Template.Spec.Containers = containers
		}

		if err := ctrl.SetControllerReference(&bar, deploy, r.Scheme); err != nil {
			logger.Error(err, "unable to set ownerReference from Bar to Deployment")
			return err
		}
		return nil
	}); err != nil {
		logger.Error(err, "unable to ensure deployment is correct")
		return ctrl.Result{}, err
	}

	// update status
	var deployment appsv1.Deployment
	var deploymentNamespacedName = client.ObjectKey{Namespace: req.Namespace, Name: bar.Spec.DeploymentName}

	if err := r.Get(ctx, deploymentNamespacedName, &deployment); err != nil {
		logger.Error(err, "unable to fetch deployment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	availableReplicas := deployment.Status.AvailableReplicas
	if availableReplicas == bar.Status.AvailableReplicas {
		return ctrl.Result{}, nil
	}

	bar.Status.AvailableReplicas = availableReplicas
	if err := r.Status().Update(ctx, &bar); err != nil {
		logger.Error(err, "unable to update Bar status")
		return ctrl.Result{}, err
	}
	r.Recorder.Eventf(&bar, corev1.EventTypeNormal, "Updated", "Update bar.status.AvailableReplicas: %d", bar.Status.AvailableReplicas)

	return ctrl.Result{}, nil
}

func (r *BarReconciler) cleanupOwnerResources(ctx context.Context, bar *samplecontrollerv1alpha1.Bar) error {
	logger := log.FromContext(ctx)
	logger.Info("finding existing Deployments for Bar resource")

	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(bar.Namespace),
		client.MatchingFields{deploymentOwnerKey: bar.Name}); err != nil {
		return err
	}
	for _, deployment := range deployments.Items {
		if deployment.Name == bar.Spec.DeploymentName {
			continue
		}

		if err := r.Delete(ctx, &deployment); err != nil {
			logger.Error(err, "failed to delete Deployment resource")
			return err
		}
		logger.Info("delete deployment resource: " + deployment.Name)
		r.Recorder.Eventf(bar, corev1.EventTypeNormal, "Deleted", "Deleted deployment %q", &deployment.Name)
	}
	return nil
}

var (
	deploymentOwnerKey = ".metadata.controller"
	apiGVStr           = samplecontrollerv1alpha1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *BarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, deploymentOwnerKey, func(rawObj client.Object) []string {
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGVStr || owner.Kind != "Bar" {
			return nil
		}
		return []string{owner.Name}

	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&samplecontrollerv1alpha1.Bar{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
