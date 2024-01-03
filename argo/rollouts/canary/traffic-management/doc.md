## 参考

https://argoproj.github.io/argo-rollouts/features/traffic-management/

## 概要

Traffic Management (データプレーンを制御してインテリジェントなルーティングルールを設定する) をすることで、
新しい Version のアプリケーションを検証している間は一部のユーザにのみ提供しリリース影響反映を制限する。

Traffic Management の実現方法はいくつかある

- percentage によってトラフィックを分散させる
- ヘッダーベースのルーティング
- トラフィックのミラーリング

## k8s デフォルトの課題

k8s デフォルトではトラフィックのミラーリングやヘッダーによるルーティングなどは行うことはできず、
サービスのセレクターに基づいた Pod のグループにルーティングするエンドポイントを提供することで負荷分散を提供する。
トラフィックの割合を制御する唯一の方法はバージョンのレプリカ数を操作することである。

## Argo Rollout での Traffic Management の有効化

Argo Rollout は利用する Service Mesh にかかわらず、仕様として canary/stable の service の設定が必要である。

```
apiVersion: argoproj.io/v1alpha1
kind: Rollout
spec:
  ...
  strategy:
    canary:
      canaryService: canary-service
      stableService: stable-service
      trafficRouting:
```

## Traffic routing with managed routes and route precedence

- Istio でサポートされている
- 重みだけでなく、argo rollout でルートを追加および管理することも可能
- 設定できるルールは header ベースおよびミラーベースのルール。これらはルータで優先順位を設定する必要がある

`managedRoutes` は Argo Rollout が Istio のリソースに対して、Route を自動で追加や削除をする。

### Traffic routing bassed on a header values for Canary

Rollout はヘッダの値によりカナリアリリースへルーティングする機能がある

### Traffic routing mirroring traffic to canary

Argo Rollout が Istio が提供する CRD の管理を自動化する。
Traffic 分割は Istio の VirtualService で定義された重みを調整して実現する。

Istio と Argo Rollout を利用する場合、２つの HTTP 宛先ルートを含む、一つの http ルートを含む VirtualService をデプロイする
２つの宛先は、Canary と Stable の宛先。

Istio は加重トラフィック分割を提供しており、Argo Rollout はどちらもオプションとしてサポートしている

1. Host-level Traffic Splitting
2. Subset-level Traffic Splitting

### Host-level Traffic Splitting

host 名または k8s Service で Stable/Canary を分割する
Istio Getting Started のサンプルはこれを利用しているっぽい。

以下のリソースのデプロイが必要

- Rollout
- Service (canary)
- Service (stable)
- VirtualService

Service は Rollout コントローラによって selector に `rollouts-pod-template-hash` という値が追加・更新される。
これは作成された replicaSet に設定された label

Argo Rollout が VirtualService や Service を書き換える

### Subset-level Traffic Splitting

2 つの DestinationRule Subset で分割する

必要なリソース

- Rollout
- Service
- VirtualService
- DestinationRule

Host-level と異なり、Service の selector はコントローラによって変更はされない
Desitination Rule の label に `rollouts-pod-template-hash` が inject される

### Host-level と Subset-level の比較

TBD
