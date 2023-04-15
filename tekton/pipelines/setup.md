### install

k8s クラスタ (minikube)

* Tekton で workspace として PersistentVolume を利用する場合、デフォルトではマルチノードに対応していないため、`CSI Hostpath Driver` を使うとよい
  * https://github.com/kubernetes/minikube/issues/12360
  * https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/
* Mac で in-cluster の insecure registry を利用するための設定をしておく
  * https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon (コードのリンクが古いので以下を参考にした)
  * https://github.com/kubernetes/minikube/blob/v1.30.1/pkg/minikube/cluster/ip.go#L128

```
// minikube start したときの VM IPが 192.168.64.xxx だったので、192.168.64.0/24 を指定した
// minikube start したときに？ 作成されている bridge100 という interface の inet が
// 192.168.64.1 だったのでこの辺りからきてる？
$ minikube -p tekton-sandbox start -n 2 --driver hyperkit --insecure-registry "192.168.64.0/24"

// volumesnapshots and csi-hostpath-driver
$ minikube -p tekton-sandbox  addons enable volumesnapshots
$ minikube -p tekton-sandbox  addons enable csi-hostpath-driver

// container registry
$ minikube -p tekton-sandbox addons enable registry
```

Tekton
```
// Pipelines
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

// Dashboard
kubectl apply --filename https://storage.googleapis.com/tekton-releases/dashboard/latest/release.yaml

// port-forward で確認
// open localhost:9097
kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097
```
