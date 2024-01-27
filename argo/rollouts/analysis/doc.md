## 参考

https://argoproj.github.io/argo-rollouts/features/analysis/#custom-resource-definitions

Argo Rollout は Progressive Delivery を推進するための分析(Analysis) を実行する方法をいくつか提供している。

## Analysis に関連するリソース

- Rollout
- AnalysisTemplate
- clusterAnalysisTemplate
- AnalysisRun
- Experiment

## Analysis 実行の種類

### Background Analysis

カナリアデプロイが進んでいる間、Analysis が実行される。終了時間は存在せず、停止か失敗するまで続行される。
Analysis が失敗すると、Rollout は Abort されて Canary の weight は 0 になり、Degraded 状態になる。

### Inline Analysis

step として Analysis を実行できる。その場合、analysis の step に到達した段階で Rollout はブロックされ、Analysis の完了を待つ。

Inline Analysis は Background と違ってずっと実行され続けるわけではない。AnalysisTemplate の `interval` と `count` によって試行回数が決定される。`interval` が指定されない場合は 1 回のみ行われる。

### Sample (canary)

```
$ kubectl create ns argo-analysis
$ kubectl -n argo-analysis apply -f rollout.yaml -f services.yaml -f analysis-template.yaml
$ kubectl argo rollouts -n argo-analysis get rollout rollouts-demo-canary-analysis --watch

// status 確認
$ kubectl argo rollouts -n argo-analysis status rollouts-demo-canary-analysis


// 更新
kubectl argo rollouts -n argo-analysis set image rollouts-demo-canary-analysis rollouts-demo=localhost/argoproj/rollouts-demo:yellow
```

### Blue/Green prePromotion/postPromotion Analysis

ReplicaSet に対するトラフィックの切り替え(active service の ReplicaSet を変更)前後 で Analysis (分析) を構成できる。
切り替え前に行うのが `prePromotionAnalysis`、切り替え後に行うのが `postPromotionAnalysis`

### Sample (Blue/Green)

```
// Mac (Apple Sillicon) 上の minikube で実行するためにローカルでビルドする
$ git clone git@github.com:argoproj/rollouts-demo.git
$ make release IMAGE_NAMESPACE=localhost/argoproj DOCKER_PUSH=false

$ minikube -p argo-sandbox image load localhost/argoproj/rollouts-demo:blue
$ minikube -p argo-sandbox image load localhost/argoproj/rollouts-demo:yellow

$ kubectl create ns argo-analysis
$ kubectl -n argo-analysis apply -f rollout.yaml -f service-active.yaml  -f service-preview.yaml -f pre-analysis-test.yaml -f post-analysis-test.yaml
$ kubectl argo rollouts -n argo-analysis get rollout rollout-bluegreen --watch

// status 確認
$ kubectl argo rollouts -n argo-analysis status rollout-bluegreen

// promote
$ kubectl argo rollouts -n argo-analysis promote rollout-bluegreen


// 更新
kubectl argo rollouts -n argo-analysis set image rollout-bluegreen rollouts-demo=localhost/argoproj/rollouts-demo:yellow
```

## memo

Rollout の失敗原因は arg-rollouts plugin の status でわかる。ただし、AnalysisRun の結果のログは job をみたりしないとわからないと思われる。
analysis の設定で、interval と count は同時に指定が必要そうだった。また、count <= failureLimit にしてしまうと絶対に成功してしまうので注意する。
https://argoproj.github.io/argo-rollouts/FAQ/#why-doesnt-my-analysisrun-end
