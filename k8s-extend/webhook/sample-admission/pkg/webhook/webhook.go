package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	whp "github.com/ginoh/sample-admission/pkg/webhook/pod"
)

func SetupWebhookWithManager(mgr ctrl.Manager) error {
	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/mutate-core-v1-pod", &webhook.Admission{Handler: whp.NewPodAnnotator(whp.PodAnnotatorClient(mgr.GetClient()))})
	hookServer.Register("/validate-core-v1-pod", &webhook.Admission{Handler: whp.NewPodValidator()})

	return nil
}
