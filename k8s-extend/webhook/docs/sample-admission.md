
Admission Webhook の実装を controller-runtime を利用した実装サンプルを試す

カスタムコントローラと一緒に実装する場合は kubebuilder を使ったが、
ここでは kubebuilder book で紹介されていた方法を利用する


### 参考

kubebuilder book
https://book.kubebuilder.io/reference/webhook-for-core-types.html

example
https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples/builtins

### 環境

controller-runtime: 
golang: 1.17.8
kubernetes: v1.23.1 (minikube)

### 実装

当初 example のコードをベースに 必要なところだけ kubebuilder v3.3.0 で生成した雛形にあわせた形で修正していたが
manifest 生成など kubebuilder (kustomize) を使った方が楽そうだったので雛形生成した後に調整した

```
kubebuilder init --domain ginoh.github.io --repo github.com/ginoh/sample-admission
kubebuilder create api --group core --version v1 --kind Pod
Create Resource [y/n]
n
Create Controller [y/n]
n
// ↑ PROJECT が更新されるだけ
kubebuilder create webhook --group core --version v1 --kind Pod --programmatic-validation --defaulting
```

生成されたコードを適宜修正

* PROJECT
  * 削除

* Makefile
  * 使いそうなところだけ適宜コピペ&修正

* hack
  * 削除

* config
  * prometheus
    * 削除
  * rbac
    * auth_proxy* leader_electoin* は削除、kustomization で role 系は今は使ってないのでコメントアウト。
  * default
    * webhoo_namespace_selector_patch追加。webhook serverデプロイ時に webhook飛ぶのを抑止
    * manager_auth_proxy は削除
  * manager
    * args を削除


* Dockerfile
  * controller, api 削除
  * pkg 追加

* main.go
  * 使わない CLI フラグの削除 (metrics とかは使わない)
  * SetupWebhookWithManager 部分の処理は削除し main.go で webhook の登録をする

* pkg
  * 追加してここに webhookのコードを置いた (マーカーは api/v1 以下のやつをコピー)
  * injectDecoder は何もしていないけどそのまま
  * mutation, validation のコードを pkg/webhook 以下において、NewXXX　で構造体のポインタを取得

* api
  * 基本的に削除
  * test suite だけ pkg/webhook に移動しつつ、中身を調整


#### build & deploy

```
export IMG=sample-webhook

make docker-build
minikube -p sandbox-controller image load --overwrite sample-webhook

make deploy
kubectl -n sample-admission-system get all

//Webhook 動作確認

//// mutate の確認
kubectl apply -f config/samples/nginx.yaml
// yq 使う
kubectl get pods nginx-pod-sample -o yaml | yq '.metadata.annotations'

//// validate　の確認

// mutate の設定を削除しておく
kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io sample-admission-mutating-webhook-configuration

kubectl patch pod nginx-pod-sample -p '{"metadata":{"annotations": {"example-mutating-admission-webhook": "bar"}}}'
Error from server (annotation example-mutating-admission-webhook did not have value "foo"): admission webhook "vpod.kb.io" denied the request: annotation example-mutating-admission-webhook did not have value "foo"

kubectl delete pods nginx-pod-sample
kubectl apply -f config/samples/nginx.yaml
Error from server (missing annotation example-mutating-admission-webhook): error when creating "config/samples/nginx.yaml": admission webhook "vpod.kb.io" denied the request: missing annotation example-mutating-admission-webhook

// uninstall
make undeploy
```

#### memo
* kubebuilder で雛形として生成される webhook マーカー中の name は

mutate が　m{{ lower .Resource.Kind }}.kb.io
validateが　v{{ lower .Resource.Kind }}.kb.io


* pod に対する webhook を登録すると、webhook Server のdeployment登録後、Podを立ち上げる際に Webhook
リクエストがきてしまう

* manager のmanifest で RunAsNonRoot指定しているため、イメージ側で root 使っていると死ぬ (雛形生成した Dockerfileなら問題ない)

* 初回の雛形は kubebuilder を使ったけど、一度作った後は使わないなら マーカーコメント削除した方がいい？

* 削除したプロジェクトは以下の内容
```
domain: ginoh.github.io
layout:
- go.kubebuilder.io/v3
projectName: sample-admission
repo: github.com/ginoh/sample-admission
resources:
- group: core
  kind: Pod
  path: k8s.io/api/core/v1
  version: v1
  webhooks:
    defaulting: true
    validation: true
    webhookVersion: v1
version: "3"
```

