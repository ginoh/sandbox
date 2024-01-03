## 参考

https://argoproj.github.io/argo-rollouts/features/canary/
https://argoproj.github.io/argo-rollouts/getting-started/istio/
https://argoproj.github.io/argo-rollouts/features/traffic-management/

## Canary デプロイの基本

https://argoproj.github.io/argo-rollouts/getting-started/

## Traffic Routing を利用した Canary デプロイ

### Cluster Setup

```
$ minikube -p argo-istio-canary start --memory=8192mb --cpus=4 --driver qemu --network socket_vmnet
```

### Istio Setup

以下の手順は Apple Silicon だと architecuture が異なり動かないイメージが使われるようだった
https://argoproj.github.io/argo-rollouts/getting-started/setup/#istio-setup

```
# Install istio
minikube addons enable istio-provisioner
minikube addons enable istio

# Label the default namespace to enable istio sidecar injection for the namespace
kubectl label namespace default istio-injection=enabled
```

Istio 公式のインストール方法比較内容を見て、今回は Helm でインストールする
https://istio.io/latest/about/faq/setup/

```
$ helm repo add istio https://istio-release.storage.googleapis.com/charts
$ helm repo update
$ helm install istio-base istio/base -n istio-system --create-namespace
$ helm install istiod istio/istiod -n istio-system --wait // Control Plane
$ helm status istiod -n istio-system
$ kubectl get deployments -n istio-system --output wide // Check


// ingress gateway
// https://istio.io/latest/docs/setup/additional-setup/gateway/
$ kubectl create namespace istio-ingress
$ helm install istio-ingressgateway istio/gateway -n istio-ingress

// 別 Terminal で tunnnel を実行。これやらないと type: LoadBalancer のリソースが pending で終わらない。
$ minikube -p argo-istio-canary tunnel
Status:
        machine: argo-istio-canary
        pid: 4245
        route: 10.96.0.0/12 -> 192.168.105.31
        minikube: Running
        services: [istio-ingress]
    errors:
                minikube: no errors
                router: no errors
                loadbalancer emulator: no errors

```

将来的には Gateway に関しては k8s の Gateway API をデフォルトにしていくらしい。

アンイストールする場合は以下
https://istio.io/latest/docs/setup/install/helm/#uninstall

### デプロイ

Istio Getting Started より以下を適用する

```
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-rollouts/master/docs/getting-started/istio/rollout.yaml
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-rollouts/master/docs/getting-started/istio/services.yaml
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-rollouts/master/docs/getting-started/istio/multipleVirtualsvc.yaml
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-rollouts/master/docs/getting-started/istio/gateway.yaml
```

ただし、`rollouts-demo` のイメージを arm64 でビルドしたものを利用するように書き換えたものを使う

```
$ kubectl apply -f rollout.yaml
$ kubectl apply -f services.yaml
$ kubectl apply -f multipleVirtualsvc.yaml
$ kubectl apply -f gateway.yaml


// そのまま実行すると以下のようなエラーがでる
// どうもSNI host は hostsの値の subsetである必要があるらしい
 2 errors occurred:
        * SNI host "reviews.bookinfo.com" is not a compatible subset of any of the virtual service hosts: [rollouts-demo-vsvc1.local]
        * SNI host "localhost" is not a compatible subset of any of the virtual service hosts: [rollouts-demo-vsvc1.local]

// 以下の変更をした
   gateways:
     - rollouts-demo-gateway
   hosts:
-    - rollouts-demo-vsvc1.local
+    - *.rollouts-demo-vsvc1.bookinfo.com
   http:
     - name: primary
       route:
@@ -24,8 +24,7 @@
     - match:
         - port: 3000
           sniHosts:
-            - reviews.bookinfo.com
-            - localhost
+            - reviews.rollouts-demo-vsvc1.bookinfo.com
       route:
         - destination:
             host: rollouts-demo-stable
@@ -43,8 +42,7 @@
   gateways:
     - rollouts-demo-gateway
   hosts:
-    #- rollouts-demo-vsvc2.local
-    - *.rollouts-demo-vsvc2.bookinfo.com
+    - rollouts-demo-vsvc2.local
   http:
     - name: secondary
       route:
@@ -62,9 +60,8 @@
     - match:
         - port: 3000
           sniHosts:
-            - reviews.rollouts-demo-vsvc2.bookinfo.com
-            # - reviews.bookinfo.com
-            # - localhost
+            - reviews.bookinfo.com
+            - localhost
       route:
         - destination:
             host: rollouts-demo-stable
```

- k8s クラスタ内での Pod へのトラフィックのルーティングに k8s svc として `rollouts-demo-canary` と `rollouts-demo-stable` を利用
- serviecs.yaml で定義
- k8s クラスタ外からのトラフィックのルーティングに istio を利用し、`VirtualService` を 2 つ指定
- multipleVirtualsvc.yaml で ViurtualService を定義している。route も定義されている。
- VirtualService で利用している gateway は gateway.yaml で定義している。

```
$ kubectl get ro
NAME DESIRED CURRENT UP-TO-DATE AVAILABLE AGE
rollouts-demo 1 1 1 38m

$ kubectl get svc
NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
kubernetes ClusterIP 10.96.0.1 <none> 443/TCP 13h
rollouts-demo-canary ClusterIP 10.106.99.216 <none> 80/TCP 4h9m
rollouts-demo-stable ClusterIP 10.108.13.76 <none> 80/TCP 4h9m

$ kubectl get virtualservice
NAME GATEWAYS HOSTS AGE
rollouts-demo-vsvc1 ["rollouts-demo-gateway"] ["*.rollouts-demo-vsvc1.bookinfo.com"] 36s
rollouts-demo-vsvc2 ["rollouts-demo-gateway"] ["*.rollouts-demo-vsvc2.bookinfo.com"] 36s

$ kubectl get gateway
NAME AGE
rollouts-demo-gateway 4h9m

$ kubectl argo rollouts get rollout rollouts-demo
Name: rollouts-demo
Namespace: default
Status: ✔ Healthy
Strategy: Canary
Step: 2/2
SetWeight: 100
ActualWeight: 100
Images: localhost/argoproj/rollouts-demo:blue (stable)
Replicas:
Desired: 1
Current: 1
Updated: 1
Ready: 1
Available: 1

NAME KIND STATUS AGE INFO
⟳ rollouts-demo Rollout ✔ Healthy 48m
└──# revision:1
 └──⧉ rollouts-demo-c5d6b64c6 ReplicaSet ✔ Healthy 10m stable
└──□ rollouts-demo-c5d6b64c6-tqzcl Pod ✔ Running 10m ready:2/2

$ kubectl argo rollouts set image rollouts-demo rollouts-demo=localhost/argoproj/rollouts-demo:yellow
$ kubectl argo rollouts get rollout rollouts-demo
kubectl argo rollouts get rollout rollouts-demo
Name:            rollouts-demo
Namespace:       default
Status:          ॥ Paused
Message:         CanaryPauseStep
Strategy:        Canary
  Step:          1/2
  SetWeight:     5
  ActualWeight:  5
Images:          localhost/argoproj/rollouts-demo:blue (stable)
                 localhost/argoproj/rollouts-demo:yellow (canary)
Replicas:
  Desired:       1
  Current:       2
  Updated:       1
  Ready:         2
  Available:     2

NAME                                       KIND        STATUS     AGE    INFO
⟳ rollouts-demo                            Rollout     ॥ Paused   10h
├──# revision:2
│  └──⧉ rollouts-demo-6877dff844           ReplicaSet  ✔ Healthy  3m47s  canary
│     └──□ rollouts-demo-6877dff844-tszpg  Pod         ✔ Running  3m47s  ready:2/2
└──# revision:1
   └──⧉ rollouts-demo-c5d6b64c6            ReplicaSet  ✔ Healthy  9h     stable
      └──□ rollouts-demo-c5d6b64c6-tqzcl   Pod         ✔ Running  9h     ready:2/2
```

この状態で、Virtual Service リソースの route に設定されている重みを確認すると、カナリアの重みが反映されるように
コントローラによって変更されている

```
$ kubectl get virtualservice rollouts-demo-vsvc1 -o yaml | yq .spec
gateways:
  - rollouts-demo-gateway
hosts:
  - '*.rollouts-demo-vsvc1.bookinfo.com'
http:
  - name: primary
    route:
      - destination:
          host: rollouts-demo-stable
          port:
            number: 15372
        weight: 95
      - destination:
          host: rollouts-demo-canary
          port:
            number: 15372
        weight: 5
tls:
  - match:
      - port: 3000
        sniHosts:
          - reviews.rollouts-demo-vsvc1.bookinfo.com
    route:
      - destination:
          host: rollouts-demo-stable
        weight: 95
      - destination:
          host: rollouts-demo-canary
        weight: 5

kubectl get virtualservice rollouts-demo-vsvc2 -o yaml | yq .spec
gateways:
  - rollouts-demo-gateway
hosts:
  - '*.rollouts-demo-vsvc2.bookinfo.com'
http:
  - name: secondary
    route:
      - destination:
          host: rollouts-demo-stable
          port:
            number: 15373
        weight: 95
      - destination:
          host: rollouts-demo-canary
          port:
            number: 15373
        weight: 5
tls:
  - match:
      - port: 3000
        sniHosts:
          - reviews.rollouts-demo-vsvc2.bookinfo.com
    route:
      - destination:
          host: rollouts-demo-stable
        weight: 95
      - destination:
          host: rollouts-demo-canary
        weight: 5

$ kubectl argo rollouts promote rollouts-demo
$ kubectl argo rollouts get rollout rollouts-demo
kubectl argo rollouts get rollout rollouts-demo
Name:            rollouts-demo
Namespace:       default
Status:          ✔ Healthy
Strategy:        Canary
  Step:          2/2
  SetWeight:     100
  ActualWeight:  100
Images:          localhost/argoproj/rollouts-demo:yellow (stable)
Replicas:
  Desired:       1
  Current:       1
  Updated:       1
  Ready:         1
  Available:     1

NAME                                       KIND        STATUS        AGE  INFO
⟳ rollouts-demo                            Rollout     ✔ Healthy     10h
├──# revision:2
│  └──⧉ rollouts-demo-6877dff844           ReplicaSet  ✔ Healthy     27m  stable
│     └──□ rollouts-demo-6877dff844-tszpg  Pod         ✔ Running     27m  ready:2/2
└──# revision:1
   └──⧉ rollouts-demo-c5d6b64c6            ReplicaSet  • ScaledDown  10h
```

canary が Scale Up されて stable になった後、もともと stable だったものが Scale Down する

## Dynamic Canary/Stable Scale

- デフォルトでは Canary を構成するレプリカは現在の Traffic の重みに対応してスケールされる

  - 全体で 4 レプリカ設定で、重み 25%の時は 1 レプリカ作られる。Stable レプリカの数はデフォルトでは変更されない。
    - Canary のレプリカ数は設定さえすれば必ずしも重みと一致させなくてもよい。
      - Canary への Traffic は流さないが、テスト目的でレプリカを起動する、Canary を 100%スケールさせて Traffic シャドーイングをするなど。
      - setCanaryScale で **スケール数**の weight や明示的な指定ができる

- Stable のレプリカ数は通常は更新中はスケール数は 100%のまま。そのため更新中のレプリカ数は最大で２倍になる (Blue/Green と同じ)
  - 設定すると動的に Stable のスケールを動的に縮小できる

## Rolling Update

- step を省略すると Rolling Update になる
