Argo Rollout 


## 準備

minikube
```
minikube -p argo-sandbox start -n 2 --driver qemu --network socket_vmnet
```

Controller
```
$ kubectl create namespace argo-rollouts
namespace/argo-rollouts created

$ kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml

customresourcedefinition.apiextensions.k8s.io/analysisruns.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/analysistemplates.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/clusteranalysistemplates.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/experiments.argoproj.io created
customresourcedefinition.apiextensions.k8s.io/rollouts.argoproj.io created
serviceaccount/argo-rollouts created
clusterrole.rbac.authorization.k8s.io/argo-rollouts created
clusterrole.rbac.authorization.k8s.io/argo-rollouts-aggregate-to-admin created
clusterrole.rbac.authorization.k8s.io/argo-rollouts-aggregate-to-edit created
clusterrole.rbac.authorization.k8s.io/argo-rollouts-aggregate-to-view created
clusterrolebinding.rbac.authorization.k8s.io/argo-rollouts created
secret/argo-rollouts-notification-secret created
service/argo-rollouts-metrics created
deployment.apps/argo-rollouts created


// plugin
$ brew install argoproj/tap/kubectl-argo-rollouts
$ kubectl argo rollouts version
kubectl-argo-rollouts: v1.4.0+e40c9fe
  BuildDate: 2023-01-09T20:26:12Z
  GitCommit: e40c9fe8a2f7fee9d8ee1c56b4c6c7b983fce135
  GitTreeState: clean
  GoVersion: go1.19.4
  Compiler: gc
  Platform: darwin/amd64
```

### デプロイ

最初に 20% のデプロイを行い、手動で Promotion した後は 自動で 20% ずつデプロイしていく設定

```
$ kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-rollouts/master/docs/getting-started/basic/rollout.yaml
$ kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-rollouts/master/docs/getting-started/basic/service.yaml

$ kubectl argo rollouts get rollout rollouts-demo --watch
・
・
・
NAME                                       KIND        STATUS              AGE  INFO
⟳ rollouts-demo                            Rollout     ✖ Degraded          20h  
└──# revision:1                                                                 
   └──⧉ rollouts-demo-687d76d795           ReplicaSet  ◌ Progressing       20h  stable
      ├──□ rollouts-demo-687d76d795-9m7dl  Pod         ✖ CrashLoopBackOff  20h  ready:0/1,restarts:90
      ├──□ rollouts-demo-687d76d795-fg4vp  Pod         ✖ CrashLoopBackOff  20h  ready:0/1,restarts:91
      ├──□ rollouts-demo-687d76d795-k9dkm  Pod         ✖ CrashLoopBackOff  20h  ready:0/1,restarts:90
      ├──□ rollouts-demo-687d76d795-qwtxh  Pod         ✖ CrashLoopBackOff  20h  ready:0/1,restarts:89
      └──□ rollouts-demo-687d76d795-zd92d  Pod         ✖ CrashLoopBackOff  20h  ready:0/1,restarts:91


// Docker Hub にあるイメージが amd64 のアーキテクチャのもののみだったので起動に失敗してるようにみえる
$ kubectl logs rollouts-demo-687d76d795-9m7dl
exec /rollouts-demo: exec format error

// Apple Silicon の Macbook で実行可能とするために arm64 アーキテクチャのイメージをビルドする
// 今回は Docker Desktop を利用してイメージをビルドすることで対応する
```
$ git clone git@github.com:argoproj/rollouts-demo.git　&& cd rollouts-demo
$ make release IMAGE_NAMESPACE=localhost/argoproj

$ minikube -p argo-sandbox image load localhost/argoproj/rollouts-demo:blue
$ minikube -p argo-sandbox image load localhost/argoproj/rollouts-demo:yellow
$ minikube -p argo-sandbox image ls
```

Rollout リソースを適用する manifest を変更

```
$ diff -u rollout.yaml rollout-update.yaml
     spec:
       containers:
       - name: rollouts-demo
-        image: argoproj/rollouts-demo:blue
+        image: localhost/argoproj/rollouts-demo:blue
         ports:
         - name: http
           containerPort: 8080

$ kubectl apply -f rollout-update.yaml
$ kubectl argo rollouts get rollout rollouts-demo --watch
Name:            rollouts-demo
Namespace:       default
Status:          ✔ Healthy
Strategy:        Canary
  Step:          8/8
  SetWeight:     100
  ActualWeight:  100
Images:          localhost/argoproj/rollouts-demo:blue (stable)
Replicas:
  Desired:       5
  Current:       5
  Updated:       5
  Ready:         5
  Available:     5

NAME                                       KIND        STATUS     AGE  INFO
⟳ rollouts-demo                            Rollout     ✔ Healthy  22s  
└──# revision:1                                                        
   └──⧉ rollouts-demo-64c8bdcb6f           ReplicaSet  ✔ Healthy  22s  stable
      ├──□ rollouts-demo-64c8bdcb6f-5wns8  Pod         ✔ Running  22s  ready:1/1
      ├──□ rollouts-demo-64c8bdcb6f-bsmbj  Pod         ✔ Running  22s  ready:1/1
      ├──□ rollouts-demo-64c8bdcb6f-jf589  Pod         ✔ Running  22s  ready:1/1
      ├──□ rollouts-demo-64c8bdcb6f-qm5km  Pod         ✔ Running  22s  ready:1/1
      └──□ rollouts-demo-64c8bdcb6f-t6v24  Pod         ✔ Running  22s  ready:1/1
```

```
$ kubectl argo rollouts set image rollouts-demo rollouts-demo=localhost/argoproj/rollouts-demo:yellow

// 20% のアップデート状態で止まる
$ kubectl argo rollouts get rollout rollouts-demo --watch
Name:            rollouts-demo
Namespace:       default
Status:          ॥ Paused
Message:         CanaryPauseStep
Strategy:        Canary
  Step:          1/8
  SetWeight:     20
  ActualWeight:  20
Images:          localhost/argoproj/rollouts-demo:blue (stable)
                 localhost/argoproj/rollouts-demo:yellow (canary)
Replicas:
  Desired:       5
  Current:       5
  Updated:       1
  Ready:         5
  Available:     5

NAME                                       KIND        STATUS     AGE  INFO
⟳ rollouts-demo                            Rollout     ॥ Paused   39s  
├──# revision:2                                                        
│  └──⧉ rollouts-demo-758cdd9dfc           ReplicaSet  ✔ Healthy  25s  canary
│     └──□ rollouts-demo-758cdd9dfc-4kdhq  Pod         ✔ Running  25s  ready:1/1
└──# revision:1                                                        
   └──⧉ rollouts-demo-64c8bdcb6f           ReplicaSet  ✔ Healthy  39s  stable
      ├──□ rollouts-demo-64c8bdcb6f-9pm5j  Pod         ✔ Running  39s  ready:1/1
      ├──□ rollouts-demo-64c8bdcb6f-rh7df  Pod         ✔ Running  39s  ready:1/1
      ├──□ rollouts-demo-64c8bdcb6f-spz7b  Pod         ✔ Running  39s  ready:1/1
      └──□ rollouts-demo-64c8bdcb6f-v74zn  Pod         ✔ Running  39s  ready:1/1

// Promote
$ kubectl argo rollouts promote rollouts-demo
$ kubectl argo rollouts get rollout rollouts-demo --watch
Name:            rollouts-demo
Namespace:       default
Status:          ✔ Healthy
Strategy:        Canary
  Step:          8/8
  SetWeight:     100
  ActualWeight:  100
Images:          localhost/argoproj/rollouts-demo:yellow (stable)
Replicas:
  Desired:       5
  Current:       5
  Updated:       5
  Ready:         5
  Available:     5

NAME                                       KIND        STATUS        AGE  INFO
⟳ rollouts-demo                            Rollout     ✔ Healthy     33m  
├──# revision:2                                                           
│  └──⧉ rollouts-demo-758cdd9dfc           ReplicaSet  ✔ Healthy     33m  stable
│     ├──□ rollouts-demo-758cdd9dfc-4kdhq  Pod         ✔ Running     33m  ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-f5nh7  Pod         ✔ Running     25m  ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-jsdp2  Pod         ✔ Running     25m  ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-p797d  Pod         ✔ Running     25m  ready:1/1
│     └──□ rollouts-demo-758cdd9dfc-8jj77  Pod         ✔ Running     25m  ready:1/1
└──# revision:1                                                           
   └──⧉ rollouts-demo-64c8bdcb6f           ReplicaSet  • ScaledDown  33m 

// Abort
$ kubectl argo rollouts set image rollouts-demo rollouts-demo=localhost/argoproj/rollouts-demo:red
$ kubectl argo rollouts get rollout rollouts-demo --watch
Name:            rollouts-demo
Name:            rollouts-demo
Namespace:       default
Status:          ॥ Paused
Message:         CanaryPauseStep
Strategy:        Canary
  Step:          1/8
  SetWeight:     20
  ActualWeight:  20
Images:          localhost/argoproj/rollouts-demo:red (canary)
                 localhost/argoproj/rollouts-demo:yellow (stable)
Replicas:
  Desired:       5
  Current:       5
  Updated:       1
  Ready:         5
  Available:     5

NAME                                       KIND        STATUS        AGE    INFO
⟳ rollouts-demo                            Rollout     ॥ Paused      75m    
├──# revision:3                                                             
│  └──⧉ rollouts-demo-566b769964           ReplicaSet  ✔ Healthy     9m54s  canary
│     └──□ rollouts-demo-566b769964-zclhh  Pod         ✔ Running     9m54s  ready:1/1
├──# revision:2                                                             
│  └──⧉ rollouts-demo-758cdd9dfc           ReplicaSet  ✔ Healthy     75m    stable
│     ├──□ rollouts-demo-758cdd9dfc-f5nh7  Pod         ✔ Running     68m    ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-jsdp2  Pod         ✔ Running     67m    ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-p797d  Pod         ✔ Running     67m    ready:1/1
│     └──□ rollouts-demo-758cdd9dfc-8jj77  Pod         ✔ Running     67m    ready:1/1
└──# revision:1                                                             
   └──⧉ rollouts-demo-64c8bdcb6f           ReplicaSet  • ScaledDown  75m 

// abort
$ kubectl argo rollouts abort rollouts-demo
$ kubectl argo rollouts get rollout rollouts-demo --watch
Name:            rollouts-demo
Name:            rollouts-demo
Namespace:       default
Status:          ✖ Degraded
Message:         RolloutAborted: Rollout aborted update to revision 3
Strategy:        Canary
  Step:          0/8
  SetWeight:     0
  ActualWeight:  0
Images:          localhost/argoproj/rollouts-demo:yellow (stable)
Replicas:
  Desired:       5
  Current:       5
  Updated:       0
  Ready:         5
  Available:     5

NAME                                       KIND        STATUS        AGE    INFO
⟳ rollouts-demo                            Rollout     ✖ Degraded    78m    
├──# revision:3                                                             
│  └──⧉ rollouts-demo-566b769964           ReplicaSet  • ScaledDown  12m    canary
├──# revision:2                                                             
│  └──⧉ rollouts-demo-758cdd9dfc           ReplicaSet  ✔ Healthy     78m    stable
│     ├──□ rollouts-demo-758cdd9dfc-f5nh7  Pod         ✔ Running     70m    ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-jsdp2  Pod         ✔ Running     70m    ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-p797d  Pod         ✔ Running     70m    ready:1/1
│     ├──□ rollouts-demo-758cdd9dfc-8jj77  Pod         ✔ Running     70m    ready:1/1
│     └──□ rollouts-demo-758cdd9dfc-6bcqc  Pod         ✔ Running     8m47s  ready:1/1
└──# revision:1                                                             
   └──⧉ rollouts-demo-64c8bdcb6f           ReplicaSet  • ScaledDown  78m    

// Degrated 状態 (Desired の状態と実際の状態が異なる)を是正するには安定版でデプロイする
// Rollout が目的の状態に達する前に安定版のマニフェストを適用するとロールバックとして検出され、Replicasetの作成は行われない
$ kubectl argo rollouts set image rollouts-demo rollouts-demo=localhost/argoproj/rollouts-demo:yellow
```