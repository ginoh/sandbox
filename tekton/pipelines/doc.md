### install

```
// k8s
// workspace を使うのに、minikube の PersistentVolume はノードが2つ以上だと権限の問題がでるので、node 1つのやつで作る
// Mac で in-cluster の insecure registry を利用するための設定をしておく
// https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon
// https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L408
$ minikube -p tekton-sandbox start --driver hyperkit --insecure-registry "10.0.0.0/24"
$ minikube -p tekton-sandbox addons enable registry

Registry を有効化したときの svc の IPが 10.xxx なのか、192.xxx　なのかがわからない。
10.xxx と 192.xxx の両方を insecure-registry に追加するか、ドメイン(registry.kube-system.svc.cluster.local)で追加したほうがいいのかもしれない

// Tekton Pipelines
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml


// Tekton Dashboard
kubectl apply --filename \
https://storage.googleapis.com/tekton-releases/dashboard/latest/tekton-dashboard-release.yaml

// port-forward
// open localhost:9097
kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097
```
