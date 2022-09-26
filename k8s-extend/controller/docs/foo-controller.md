実践入門 Kubernetes カスタムコントローラーへの道 第5章 を試し
見返すためのメモ

### 実装メモ

### 初期化と API追加
```
kb init --domain ginoh.github.io --repo github.com/ginoh/foo-controller
kb create api --group samplecontroller --version v1alpha1 --kind Foo
```

`domain` に k8s.io は使えない。ビルド 時に以下のようなエラーにあるようにプロテクトされている

```
"unable to install CRDs onto control plane: unable to create CRD instances: unable to create CRD \"foos.samplecontroller.k8s.io\": CustomResourceDefinition.apiextensions.k8s.io \"foos.samplecontroller.k8s.io\" is invalid: metadata.annotations[api-approved.kubernetes.io]: Required value: protected groups must have approval annotation \"api-approved.kubernetes.io\", see https://github.com/kubernetes/enhancements/pull/1111",
```

FooSpec int32 ではなく *int32  の理由
deployment の replicasが pointer で定義されている？
=> pointer パッケージを使うとよいかも？


Reconcile 処理が動作するのは Event が発生し Request が WorkQueue に入ったタイミング

* Reconcileの実装で リソースのGetがエラーの場合に errors.IsNotFound() を利用するか、client.IgnoreNotFound() を利用するかは
NotFoundとそうでないエラーでやりたい処理が違う時に使い分けるとよさそう

* CR内で別リソースを指定していて、依存関係を持つような場合、CRの指定を変更すると元々依存関係があった古いリソースがゴミとして残るため、クリーンアップ処理をちゃんとしておく

* client.MatchingFields は今は関数ではない
```
client.MatchingFields{deploymentOwnerKey: foo.Name}
```

* createOrUpdate の callback処理において、selectorは不変なため新規作成時のみの処理
このあたりを参考にするといいかも
https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/controller/controllerutil#CreateOrUpdate

### webhook (mutating, validating)
```
kubebuilder create webhook --group samplecontroller --version v1alpha1 --kind Foo --programmatic-validation --defaulting
```
* 自動生成 では webhookの実装は api/<version> の下にコードが生成される
* デフォルト値設定の mutating は Default を実装
* validatingは validateCreate, validateUpdate, vaidateDelete の実装


#### conversion webhook

apiversion  の追加
```
kb create api --group samplecontroller --version v1beta1 --kind Foo
```
* リソースは作るけど、コントローラは作らない
* storageversionのマーカーを v1alpha1 側につける、これにより etcd への保存 versionは そのバージョンが利用される
* api/v1alpha1, api/v1beta に foo_conversion.go を実装
  * v1alpha1 は Hub関数を宣言するだけ
* config/crd/kustomization.yaml の patchのコメントアウトを外す

deploy 後、sampleをapply => v1beta1 で get して確認する
```
kc apply -f config/samples/samplecontroller_v1alpha1_foo.yaml
```

conversion webhook に問題があると get できなくなるっぽい
```
Error from server: conversion webhook for samplecontroller.ginoh.github.io/v1alpha1, Kind=Foo failed: the server could not find the requested resource
```

### other

controller の importに大体使いそうなやつ

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
