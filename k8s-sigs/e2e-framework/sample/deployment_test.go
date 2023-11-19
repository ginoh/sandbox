package sample_test

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

type key int

const (
	sampleKey key = iota
)

func TestDeployment(t *testing.T) {
	deploymentFeature := features.New("appsv1/deployment").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := newDeployment(c.Namespace(), "test-deployment", 1)
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			if err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue),
				wait.WithTimeout(time.Minute*5)); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("deployment creation", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var dep appsv1.Deployment
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Get(ctx, "test-deployment", c.Namespace(), &dep); err != nil {
				t.Fatal(err)
			}
			t.Logf("deployment found: %s", dep.Name)
			return context.WithValue(ctx, sampleKey, &dep)
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			cv := ctx.Value(sampleKey)
			if cv == nil {
				return ctx
			}
			dep := cv.(*appsv1.Deployment)
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Delete(ctx, dep); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testEnv.Test(t, deploymentFeature)
}

func newDeployment(namespace string, name string, replicas int32) *appsv1.Deployment {
	labels := map[string]string{"app": "test-app"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "nginx"}}},
			},
		},
	}
}
