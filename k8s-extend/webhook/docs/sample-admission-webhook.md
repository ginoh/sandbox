Admission Webhook の実装を kubebuilder を利用して試す

k8s Core グループのリソースに対する Webhook の作成は kubebuilder でサポートしていないということだったが、controller-runtime でサポートしているようで、少し修正すれば kubebuilder の仕組みにのっかれそうだったので、以前試したものを kubebuilder を使って作ってみる

### 参考

[Kubebuilder で Core Resource の Admission Webhook を作る](https://tech.griphone.co.jp/2021/12/12/kubebuilder-coreresource-webhook/)

### 環境

golang: 1.19.1
kubernetes: v1.24.3 (minikube)
kubebuilder: 3.7.0

###  今回作るもの
* Pod 作成・更新時に annotation をつける (mutating)
* Pod 作成・更新時に 名前の命名規則をチェックをする (validating)
  * ただし、validation を無視する annotation があると validation を行わない

### 準備

```
minikube -p sample-admission-webhook start --driver hyperkit --insecure-registry "10.0.0.0/24"

kubebuilder init --domain ginoh.github.io --repo github.com/ginoh/sample-admission-webhook
kubebuilder create api --group core --version v1 --resource=false --controller=false --kind Pod
kubebuilder create webhook --group core --version v1 --kind Pod --programmatic-validation --defaulting
```

create api 時に、
* `--resource=false` で api ディレクトリは作成されなくなる
* `--controller=false` で controllers ディレクトリは作成されなくなる
* 結果的に PROJECT が変更されるだけ

以下が変更差分
```
--- a/k8s-extend/webhook/sample-admission-webhook/PROJECT
+++ b/k8s-extend/webhook/sample-admission-webhook/PROJECT
@@ -3,4 +3,9 @@ layout:
 - go.kubebuilder.io/v3
 projectName: sample-admission-webhook
 repo: github.com/ginoh/sample-admission-webhook
+resources:
+- group: core
+  kind: Pod
+  path: k8s.io/api/core/v1
+  version: v1
 version: "3"
```

create webhook 実行時に、
```
kubebuilder create webhook --group core --version v1 --kind Pod --programmatic-validation --defaulting
Writing kustomize manifests for you to edit...
Writing scaffold for you to edit...
api/v1/pod_webhook.go
Update dependencies:
$ go mod tidy
Running make:
$ make generate
・
・ (省略)
・
Next: implement your new Webhook and generate the manifests with:
$ make manifests
```
* PROJECT 修正
* main.go に Setupコード追加
* 以下が追加
  * api/
  * config/certmanager/
  * config/default/manager_webhook_patch.yaml
  * config/default/webhookcainjection_patch.yaml
  * config/webhook/

### 実装

#### Webhook を動作させるための修正

Dockerfile

* 今回、controllers ディレクトリは作成していないので、Dockerfile 中で COPY instruction でディレクトリをコピーしている部分をコメントアウトしておく

pod_webhook.go の修正

* 以下の Defaulter, Validator の記述の部分を、それぞれ CustomDefulter, CustomValidator を使うようにし、また設定する構造体を定義しておく
  * Pod 構造体のままでもいいが、 Pod リソースの構造体とまぎらわしいため、参考 Web ページのようにPodWebhook という名前にしておく
  * PodWebhook という名前にするにあたり、レシーバの名前を変更するのと、メソッドのsignature が異なっているので修正する
    * https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/webhook/admission#CustomDefaulter
    * https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/webhook/admission#CustomValidator
  * SetupWebhookWithManager() は利用しないので削除する
* +kubebuilder:webhook マーカーの name の core部分を削除する
  * この名前は webhook の configurationリソース作る時のパスになるが、controller-runtime では実装的に GVK を取得すると Group は 空文字列になりサーバとしては core がないパスが登録されるため、それに対応しておく。ちなみにどんなパスが登録されているかは、Pod起動時のログをみればわかる。
  * https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/builder/webhook.go#L208
```
type PodWebhook struct{}
var _ webhook.Defaulter = &Pod{}
// var _ webhook.CustomDefaulter = &PodWebhook{}
・
・
・
var _ webhook.Validator = &Pod{}
// var _ webhook.CustomValidator = &PodWebhook{}
```

main.go の修正

webhook の以下のセットアップの自動生成コード部分を置き換える
```
	if err = (&corev1.Pod{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Pod")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder
```

変更後
```
	if err = ctrl.NewWebhookManagedBy(mgr).
		WithDefaulter(&v1.PodWebhook{}).
		WithValidator(&v1.PodWebhook{}).
		For(&corev1.Pod{}).
		Complete(); err != nil {
		setupLog.Error(err, "unable to create controller", "webhook", "Pod")
		os.Exit(1)
	}
```

config/default/kustomization.yaml の修正
* 以下コメントアウト
  * ../crd
* 以下コメントアウト外す
  * ../webhook
  * ../certmanager
  * manager_webhook_patch.yaml
  * webhookcainjection_patch.yaml
  * vars
* 後述する以下を patchesStrategicMerge に追加
  * webhook_namespace_selector_patch.yaml

config/rbac/kustomization.yaml の修正
* 以下コメントアウト
  * role.yaml

k8s API Server にアクセスする Controller が存在しないのもあり、Role を作成することがない。そのため、role.yaml はコメントアウトする

config/default/webhook_namespace_selector_patch.yaml を追加

Pod リソースに対して Webhook を実行するが、Webhook サーバの Pod をデプロイ時に Webhook を実行しようとして失敗するので、namespace_selector を利用して除外する



config/manager/manager.yaml
* imagePullPolicy: IfNotPresent を追加


api/v1/webhook_suite_test.go の修正

API Resource 新しく作ったわけではないので、schema周りを修正
```
  import (
    ・・・
    clientgoscheme "k8s.io/client-go/kubernetes/scheme"
    ・・・
  )

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
```
webhook の Setup を修正
```
// 変更前
err = (&Pod{}).SetupWebhookWithManager(mgr)

// 変更後
err = ctrl.NewWebhookManagedBy(mgr).
  WithDefaulter(&PodWebhook{}).
  WithValidator(&PodWebhook{}).
  For(&corev1.Pod{}).
  Complete()
```



#### 動作確認

cert-manager インストール
```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
```

webhook サーバのビルド & インストール
```
IMG=sample-webhook:latest make docker-build
minikube -p sample-admission-webhook image load sample-webhook:latest
IMG=sample-webhook:latest make deploy

namespace/sample-admission-webhook-system created
serviceaccount/sample-admission-webhook-controller-manager created
role.rbac.authorization.k8s.io/sample-admission-webhook-leader-election-role created
clusterrole.rbac.authorization.k8s.io/sample-admission-webhook-metrics-reader created
clusterrole.rbac.authorization.k8s.io/sample-admission-webhook-proxy-role created
rolebinding.rbac.authorization.k8s.io/sample-admission-webhook-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/sample-admission-webhook-manager-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/sample-admission-webhook-proxy-rolebinding created
service/sample-admission-webhook-controller-manager-metrics-service created
service/sample-admission-webhook-webhook-service created
deployment.apps/sample-admission-webhook-controller-manager created
certificate.cert-manager.io/sample-admission-webhook-serving-cert created
issuer.cert-manager.io/sample-admission-webhook-selfsigned-issuer created
mutatingwebhookconfiguration.admissionregistration.k8s.io/sample-admission-webhook-mutating-webhook-configuration created
validatingwebhookconfiguration.admissionregistration.k8s.io/sample-admission-webhook-validating-webhook-configuration created
```

```
kubectl -n sample-admission-webhook-system logs -f <webhook pod> -c manager

// 上記の後、別 Terminal で
kubectl create ns pod-webhook-test
kubectl -n pod-webhook-test apply -f config/samples/sample-envoy.yaml

// webhook サーバの Pod にこんな感じのログがでる
1.6655964855662575e+09  DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/mutate--v1-pod", "UID": "9333ccfa-7051-4a25-a624-ad8743857b93", "kind": "/v1, Kind=Pod", "resource": {"group":"","version":"v1","resource":"pods"}}
1.6655964855733461e+09  INFO    pod-resource    default {"name": "sample-envoy"}
1.6655964855789735e+09  DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/mutate--v1-pod", "code": 200, "reason": "", "UID": "9333ccfa-7051-4a25-a624-ad8743857b93", "allowed": true}
1.6655964855905197e+09  DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/validate--v1-pod", "UID": "2ad9da9e-dec0-4d30-9609-375727347b4c", "kind": "/v1, Kind=Pod", "resource": {"group":"","version":"v1","resource":"pods"}}
1.6655964855914538e+09  INFO    pod-resource    validate create {"name": "sample-envoy"}
1.6655964855917501e+09  DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/validate--v1-pod", "code": 200, "reason": "", "UID": "2ad9da9e-dec0-4d30-9609-375727347b4c", "allowed": true}
```

#### CustomDefaulter の実装

api/v1/pod_webhook.go

* ランダムの10文字の文字列を値にもつ annotation を追加している
* obj を pod の構造体で型アサーションして利用する



#### CustomValidator の実装

api/v1/pod_webhook.go

* エラー に関しては 以下のようにするとよさそう
  * 個々のエラーは field.Error 型を利用
  * エラーを field.ErrorList にまとめる
  * 最終的には k8s.io/apimachinery/pkg/api/errors パッケージで apierrors.StatusError を返却

### テスト実装

api/v1/webhook_suite_test.go

* test 用の Pod リソースを反映するための namespace を作れるようにする
```
ns := &corev1.Namespace{}
ns.Name = "test"
err = k8sClient.Create(context.Background(), ns)
Expect(err).NotTo(HaveOccurred())
```
* BeforeEach で テスト用の Pod リソースを削除するようにする
  * test 用 namespace をわける、namespaceごと削除するといった方法でもいいかもしれない

#### 動作確認
```
kubectl -n pod-webhook-test  apply -f config/samples/sample-envoy.yaml
pod/sample-envoy created

kubectl -n pod-webhook-test get pods sample-envoy -o yaml | yq .metadata.annotations
kubectl.kubernetes.io/last-applied-configuration: |
  {"apiVersion":"v1","kind":"Pod","metadata":{"annotations":{},"labels":{"run":"envoy"},"name":"sample-envoy","namespace":"pod-webhook-test"},"spec":{"containers":[{"image":"envoyproxy/envoy-dev","name":"envoy","resources":{}}]}}
sample-admission-webhook/test-key: 3b1eaf2e5a

kubectl -n pod-webhook-test apply -f config/samples/invalid-envoy.yaml              
The Pod "invalid-envoy" is invalid: metadata.name: Invalid value: "invalid-envoy": name must be prefixed with sample-

// invalid-envoy.yaml の annotations に "sample-admission-webhook/ignore": "true" を追加すると Pod を作成可能
```



