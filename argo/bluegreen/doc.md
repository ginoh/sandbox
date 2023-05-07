## 参考

https://argoproj.github.io/argo-rollouts/features/bluegreen/

## 概要

Blue/Green Deploy を利用することで異なる Version のアプリケーションの実行時間を短縮する。
ロールアウトの仕様では Active なサービスへの参照とオプションとして preview サービスの参照を持つ。

Rollout が更新され、新しい ReplicaSet が作成されると、Service の selector を編集し新しい Version へ向くようにする。

## Blue/Green の動作シーケンス

1. 完全に promote された定常状態からはじまるとすると、`activeService` と `previewService` は同一の ReplicaSet (Revision 1) にポイントされている。
2. ユーザが Pod Template を更新して更新を開始する
3. Size が 0 の ReplicaSet を Revision 2 として作成する
4. `previewService` が Revision 2 の ReplicaSet を指すように変更される。`activeService` は変更されない。
5. `previewService` は `spec.replicas` の値もしくは設定されていれば `previewReplicaCount` が利用される
6. Revision 2 の ReplicaSet Pod が利用可能になると、 `prePromotionAnalysis` が開始される
7. prePromotion が成功すると `autoPromotionEnabled` が false か `autoPromotionSeconds` が 0 以外のときは Pause する。
8. Rollout をユーザが手動で再開するか、`autoPromotionSeconds` を超えると自動的に再開される。
9. `previewReplicaCount` を利用していた場合、`spec.replicas` の値にスケールする
10. `activeService` を更新して Revision 2 の ReplicaSet を指すように Promote する。この時点で Revision 1 の ReplicaSet を指すサービスはなくなる。
11. `postPromotionAnalysis` を開始する。
12. `postPromotionAnalysis` が成功すると Revision 2 の ReplicaSet が Stable としてマークされる。
13. `scaleDownDelaySeconds (default 30sec)` 待機した後、Revision 1 の ReplicaSet が縮小される。

## サンプル実行

```
$ kubectl -n argo-bluegreen apply -f rollout.yaml -f service-active.yaml -f service-preview.yaml
$ kubectl argo rollouts get -n argo-bluegreen rollout rollout-bluegreen --watch
$ kubectl argo rollouts -n argo-bluegreen set image rollout-bluegreen rollouts-demo=localhost/argoproj/rollouts-demo:yellow
$ kubectl argo rollouts -n argo-bluegreen get rollout rollout-bluegreen --watch
$ kubectl argo rollouts -n argo-bluegreen promote rollout-bluegreen
```
