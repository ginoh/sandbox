package v1beta1

import (
	samplecontrollerv1alpha1 "github.com/ginoh/foo-controller/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Foo to the Hub version (v1alpha1)
func (src *Foo) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*samplecontrollerv1alpha1.Foo)

	// meta
	dst.ObjectMeta = src.ObjectMeta
	// spec
	dst.Spec.DeploymentName = src.Spec.DeploymentName
	dst.Spec.Replicas = src.Spec.Replicas
	// status
	dst.Status.AvailableReplicas = src.Status.AvailableReplicas

	return nil
}

// ConvertFrom converts from the Hub version (v1alpha1) to this version
func (dst *Foo) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*samplecontrollerv1alpha1.Foo)

	dst.Spec.Foo = src.Spec.DeploymentName

	// meta
	dst.ObjectMeta = src.ObjectMeta
	// spec
	dst.Spec.DeploymentName = src.Spec.DeploymentName
	dst.Spec.Replicas = src.Spec.Replicas
	// status
	dst.Status.AvailableReplicas = src.Status.AvailableReplicas

	return nil
}
