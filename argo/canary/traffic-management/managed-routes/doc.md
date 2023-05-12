## routing based on a header value

```
$ kubectl -n managed-routes apply -f virtualsvc.yaml -f rollout.yaml -f services.yaml -f gateway.yaml
$ kubectl argo rollouts -n managed-routes get rollout rollouts-managed-routes

// ingressgateway の External IP を確認
$ kubectl -n istio-ingress get svc
NAME                   TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)                                      AGE
istio-ingressgateway   LoadBalancer   10.102.164.188   10.102.164.188   15021:30107/TCP,80:30861/TCP,443:30596/TCP   2d5h

// curl で確認
$ curl --resolve istio-canary-header-routing.com:80:10.102.164.188 http://istio-canary-header-routing.com/color
"blue"

// image の更新をする
$ kubectl argo rollouts -n managed-routes set image rollouts-managed-routes rollouts-demo=localhost/argoproj/rollouts-demo:yellow

// Canary リリースの確認
$ kubectl argo rollouts -n managed-routes get rollout rollouts-managed-routes --watch
Name:            rollouts-managed-routes
Namespace:       managed-routes
Status:          ॥ Paused
Message:         CanaryPauseStep
Strategy:        Canary
  Step:          2/4
  SetWeight:     25
  ActualWeight:  25
Images:          localhost/argoproj/rollouts-demo:blue (stable)
                 localhost/argoproj/rollouts-demo:yellow (canary)
Replicas:
  Desired:       1
  Current:       2
  Updated:       1
  Ready:         2
  Available:     2

NAME                                                 KIND        STATUS     AGE    INFO
⟳ rollouts-managed-routes                            Rollout     ॥ Paused   3h31m
├──# revision:2
│  └──⧉ rollouts-managed-routes-945b56b66            ReplicaSet  ✔ Healthy  19s    canary
│     └──□ rollouts-managed-routes-945b56b66-bkgkh   Pod         ✔ Running  19s    ready:2/2
└──# revision:1
   └──⧉ rollouts-managed-routes-547446cf7d           ReplicaSet  ✔ Healthy  3h31m  stable
      └──□ rollouts-managed-routes-547446cf7d-882ts  Pod         ✔ Running  3h31m  ready:2/2

// curl で確認
$ curl --resolve istio-canary-header-routing.com:80:10.102.164.188 http://istio-canary-header-routing.com/color
"blue"

$  curl --resolve istio-canary-header-routing.com:80:10.102.164.188 http://istio-canary-header-routing.com/color -H "Custom-Header1:Test"
"yellow"

// promote
$ kubectl argo rollouts -n managed-routes promote rollouts-managed-routes
```

`setHeaderRoutes` の step に到達した時点で `VirtualService` リソースは以下のようにルートが追加される

```
・
・
・
spec:
  gateways:
  - rollouts-demo-gateway
  hosts:
  - istio-canary-header-routing.com
  http:
  - match:
    - headers:
        Custom-Header1:
          exact: Test
    - headers:
        Custom-Header2:
          prefix: Test
    - headers:
        Custom-Header3:
          regex: Test(.*)
    name: set-header-1
    route:
    - destination:
        host: rollouts-demo-canary
      weight: 100
  - name: primary
    route:
    - destination:
        host: rollouts-demo-stable
      weight: 100
    - destination:
        host: rollouts-demo-canary
      weight: 0
```

この後の Step で `setWeight: 50` などとした場合は、primary route の `rollouts-demo-canary` の weight が変更される。

step が最後まで完了し、canary が stable に移行すると以下のように `VirtualService` は最初の状態と同じようになっていた。

```
spec:
  gateways:
  - rollouts-demo-gateway
  hosts:
  - istio-canary-header-routing.com
  http:
  - name: primary
    route:
    - destination:
        host: rollouts-demo-stable
      weight: 100
    - destination:
        host: rollouts-demo-canary
      weight: 0
```
