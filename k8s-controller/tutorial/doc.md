kubebuilder を使ったカスタムコントローラ作成の Tutorial に関して見返すようにメモ

### 参考

以下のドキュメントに従って試したことをメモする

つくって学ぶKubebuilder
https://zoetrope.github.io/kubebuilder-training/

標準の controller を参考に

https://github.com/kubernetes/kubernetes/tree/master/pkg/controller

### 環境

* kubebuilder (v3.4.0)
* kubernetes (v1.23.1, minikube)
* go (1.17.8, goenv)

### カスタムコントローラ作成に利用するツール

* kubebuilder
  * カスタムコントローラのプロジェクトの雛形を生成するツール
* controller-tools
  * Goのソースコードからマニフェストを生成するツール
* controller-runtime
  * カスタムコントローラを実装するためのフレームワーク・ライブラリ

### kubebuilder を利用した開発
#### 初期化
```
kubebuilder init --domain ginoh.github.io --repo github.com/ginoh/markdown-view
```
`domain` => CRD group

`repo` => go.mod のモジュール名

* main.go は entrypointとなるコード
* config ディレクトリはデプロイに利用する manifest群
  * manager はコントローラの manifest
  * default は config以下の manfestをまとめて利用するための設定

#### API/Webhook の雛形生成
```
kubebuilder create api --group view --version v1 --kind MarkdownView
```
* api/v1
  * カスタムリソースの Goの構造体表現
* controllers
  * メインロジック
* config/crd
  * go 構造体から自動で生成される

```
kubebuilder create webhook --group view --version v1 --kind MarkdownView --programmatic-validation --defaulting
```
* api/v1
  * webhook を実装していく。
* config/certmanager
  * Adminssion Webhook機能を利用するのに証明書が必要なため、cert-managerを利用して証明書発行するためのカスタムリソース
* config/webhook
  * webhookを利用するための manifest

開発の流れとしては

go による実装 => manifestを生成 => 生成した manifest を適用

cert-manager をインストールしておく
```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
```

#### コントローラのデプロイ

イメージのビルドと minikube へのイメージのロード

デフォルトでは、 `controller:latest` というイメージ名でビルドされる
`IMG` 変数を指定することで変更することが可能
```
// minikube でクラスタ作成
minikube -p sandbox-controller --driver hyperkit start -n 3

// IMG=XXXX make docker-build
make docker-build

minikube -p sandbox-controller image load --overwrite controller:latest
```
latest を使う場合は config/manager/manager.yaml  に imagePullPolicy を設定しておくとよい

minikube で image load するときに、同じ名前の場合は `overwrite`を指定が必要？

```
// CRD インストール
make install

// コントローラのデプロイ (CRD含む)
make deploy

// e.g.
kubectl -n markdown-view-system logs markdown-view-controller-manager-6698746f45-kvxt7 -c manager -f
```

実装時に実行するもの
* 実装が変わる => make docker-build, image load
* CRD に変更がある => make install
* CRD 以外の manifest に変更がある => make deploy

各変更適用後、kubectl rollout で restart する

#### controller-tools と controller-runtime
controller-tools については以下を参照
https://zoetrope.github.io/kubebuilder-training/controller-tools/

controller-runtime  については以下を参照
https://zoetrope.github.io/kubebuilder-training/controller-runtime/

削除リクエストを投げる前に、同じ名前のリソースが作成される可能性があるため
Precondition で削除対象を指定できる

#### Reconcileの実装

カスタムコントローラのロジックとして reconcile.Reconciler interface を実装する
```
type Reconciler interface {
    Reconcile(context.Context, Request) (Result, error)
}
```
Reconcileの実行タイミング
* コントローラーの扱うリソースが作成、更新、削除されたとき
* Reconcileに失敗してリクエストが再度キューに積まれたとき
* コントローラーの起動時
* 外部イベントが発生したとき
* キャッシュを再同期するとき(デフォルトでは10時間に1回)

Reconcile処理はデフォルトでは1秒間に10回以上実行されないように制限されている
イベントが高い頻度で発生する場合は、Reconciliation Loopを並列実行するように設定可能

監視対象の制御
* NewControllerManagedBy 関数を利用
 * For => 監視対象
 * Owns => 生成するリソースに何らかの変更が発生した際にReconcileが呼び出されるようにな


Reconciler に Recorder を追加
=> https://zoetrope.github.io/kubebuilder-training/controller-runtime/manager.html を参考に mainの処理,manifest更新などをする

#### リソースの削除
* ownerReferenceによるガベージコレクション
* Finalizer

ownerReference
親リソースが削除されると、そのリソースの子リソースもガベージコレクションにより自動的に削除されるという仕組み
`.metadata.ownerReferences` で親子関係を表現。`controllerutil.SetControllerReference` で設定できる
異なるnamespaceのリソースをownerにしたり、cluster-scopedリソースのownerにnamespace-scopedリソースを指定することはできない

`controllerutil.SetOwnerReference` も存在
違いは SetControllerReferenceは、1つのリソースに1つのオーナーのみしか指定できず、controllerフィールドとblockOwnerDeletionフィールドにtrueが指定されているため子リソースが削除されるまで親リソースの削除がブロックされる。
一方のSetOwnerReferenceは1つのリソースに複数のオーナーを指定でき、子リソースの削除はブロックされない

Finalizer
直接の親ではないリソースを削除したいケースや、Kubernetesで管理していない外部のリソースなどを削除
* リソースのfinalizersフィールドにFinalizerの名前を指定
* リソースの削除をしようとすると削除はされず、deletionTimestampが付与されるだけ
* カスタムコントローラーはdeletionTimestampが付与されていることを発見すると、そのリソースに関連するリソースを削除し、その後にfinalizersフィールドを削除

#### Controller のテスト

controller-runtime は envtest というパッケージを提供していて、コントローラや Webhook の簡易的なテストを実施できる
envtest => etcd と kube-apiserver を立ち上げてテスト用の環境を構築
環境変数 `USE_EXISTING_CLUSTER` を指定すると既存の k8s クラスタを利用したテストも可能

* Envtestでは、etcdとkube-apiserverのみを立ち上げており、controller-managerやschedulerは動いていない
* DeploymentやCronJobリソースを作成しても、Podは作成されない
* controller-genが生成するテストコードでは、Ginkgoというテストフレームワークを利用している

具体的な実装手順
* envtest.Environment でテスト用の環境設定
* testEnv.Start() を呼び出すと etcd と kube-apiserver が起動する
* テスト終了時には etcd と kubeapiserver を終了するように testEnv.Stop() を呼び出す
* Reconcile処理はテストコードとは非同期に動くため、Eventually関数を利用してリソースが作成できるまで待つ

```
make test
```
