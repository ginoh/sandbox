## 準備 (kubebuilder のアップデートしておく)
```
// k8s cluster
minikube -p sample-bar-contrller start -n 3 --driver hyperkit --insecure-registry "10.0.0.0/24"

// cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml

curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
// 自分の好きなところへ移動
chmod +x kubebuilder && mv kubebuilder ~/bin/
```
## 初期化/起動確認
```
kubebuilder init --domain ginoh.github.io --repo github.com/ginoh/bar-controller
kubebuilder create api --group samplecontroller --version v1alpha1 --kind Bar
```
* api/v1
  * カスタムリソースの Goの構造体表現
* controllers
  * メインロジック
* config/crd
  * go 構造体から自動で生成される


manager.yaml で imagePullPolicyを設定
```
containers:
- command:
    - /manager
    args:
    - --leader-elect
    image: controller:latest
    imagePullPolicy: IfNotPresent
    name: manager
```

以下で Controller をデプロイして起動
```
IMG=bar-controller:dev make docker-build
minikube -p sample-bar-contrller image load --overwrite bar-controller:dev
make install
IMG=bar-controller:dev make deploy
kubectl -n bar-controller-system logs bar-controller-controller-manager-65c7cdb47c-dk49m -c manager -f
```
この時点では起動したことがわかるだけ

## 実装

* kubernetes/sample-controller リポジトリと同等のカスタムリソース・コントローラ
* deployment を管理する bar リソースを作成
### 仕様の定義

api/v1alpha1/bar_types.go を変更
* Spec を変更
  * +optional のマーカーをつけると省略可能なフィールドになる
  * +kubebuilder:validation:Optional の指定でも同様なことが可能だが、こちらは packageレベルでも適用できる。その場合、マーカーなしのデフォルトの挙動が optional になる
  * フィールドのタグに omitempty をつけていると自動で optional なフィールドになる
  * optional なフィールドは指定しない場合の挙動がポインタ型の場合は nil がそうでない場合はゼロ値
* Status を変更
* Bar　にマーカー追加
  * 基本的には変更しない。カラム表示のマーカーを追加する。

### コントローラ実装

controller の修正
* controller に マーカーをつけることで rbac リソースを生成

役割
* manager
* client


main.go
* client を生成、 標準リソースを扱えるように scheme を追加してしている
* status の更新は status 更新用クライアントを使う

Reconcile 実装
* 戻り値は
  * 処理成功 & Requeue なし => return ctrl.Result{}, nil
  * 処理失敗 & Requeue あり => return ctrl.Result{}, err
  * 明示的に Requeue => ctrl.Result{Requeue: true}, nil

* SetControllerReference でリソースの親子関係を設定
  * 似たようなものに、SetOwnerReference があるが、こちらは一つのリソースに複数の Owner が設定が可能
* Status 更新
  * Event を記録するために 構造体に Record を持たせた上で、 main.go で渡す
  * rbac マーカーの追加 (controller 動作のための rbac は controller ロジックのところに追加)

* クリーンアップ処理
  * カスタムリソースが管理している(ことになっている) Deployment を取得し、現在の Spec と比較し異なる場合は削除する
  * リソース取得するのに特定のフィールドでフィルタリングしたい場合はあらかじめ index をはっておく必要があり、setupWithManager で行う

import 関連でよく使いそうなもの
```
appsv1 "k8s.io/api/apps/v1"
corev1 "k8s.io/api/core/v1"
(metav1 "k8s.io/apimachinery/pkg/apis/meta/v1")
"k8s.io/client-go/tools/record"
("k8s.io/utils/pointer")
("sigs.k8s.io/controller-runtime/pkg/client/apiutil")
("sigs.k8s.io/controller-runtime/pkg/controller/controllerutil")

// server-side apply
(appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1")
(corev1apply "k8s.io/client-go/applyconfigurations/core/v1")
(metav1apply "k8s.io/client-go/applyconfigurations/meta/v1")
```

### 動作確認
```
IMG=bar-controller:dev make docker-build
make install
minikube -p sample-bar-contrller image load --overwrite bar-controller:dev
IMG=bar-controller:dev make deploy

kubectl create ns sample-bar
kubectl -n sample-bar apply -f config/samples/samplecontroller_v1alpha1_bar.yaml
```

### コントローラ周辺詳細知識

client-go と Custom Controller を組み合わせて処理を行うときの仕組み

参考：
* https://github.com/kubernetes/sample-controller/blob/master/docs/controller-client-go.md
* 実践入門 Kubernetes カスタムコントローラーへの道 2章

 
client-go のコンポーネントの役割
1 オブジェクトのイベントを監視して、オブジェクトのデータをインメモリのキャッシュに保存する
  * コントローラがオブジェクトのデータを取得するときはこのキャッシュからデータを取得する
2 Event Handler を経由して、オブジェクトをコントローラの WorkQueue に送る
  * WorkQueue は Reconcile 処理をするためのアイテムをためておく Queue

1 の処理の関連コンポーネントとして以下がある
* Reflector ・・・k8s API Server に対してオブジェクトを監視
* Delta FIFO・・・Reflector が Event を検知すると更新されたオブジェクトが入る
* Informor・・・FIFO からオブジェクトを pop して、indexer に追加
* Indexer・・・オブジェクトを Store に保存
* Store (in-memory-cache)・・・オブジェクトが保存されている
* Lister・・・オブジェクトのデータ取得は、Lister が Store からデータ取得して行われる (indexer経由)


2 の処理の関連コンポーネントとして以下がある
* Reflector・・・ 1と共通
* Delta FIFO ・・・1と共通
* Informor・・・1と共通だが、Event Handler を呼び出すことで、pop したオブジェクトの key を取得し、 コントローラの Work Queue に key を追加

controller のコンポーネントの主な役割
1 Work Queue からデータを取り出し、 Reconcile 処理を行う


* Event Handler・・・Informor によって呼び出されるコールバック関数
* Work Queue・・・Reconcile 処理を行うオブジェクトのキー　を保持する Queue。タイプの違う Queue がある。
* Process Item・・・Work Queue からのアイテムを処理する関数(コントローラのロジック)で1つ以上存在する。通常、key に対応するオブジェクトを取得して利用する。


### webhook 実装