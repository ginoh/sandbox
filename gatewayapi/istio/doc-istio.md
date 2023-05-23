## 参考

Gateway API の理解
https://thinkit.co.jp/article/19625 (前編)

- https://www.youtube.com/watch?v=lCRuzWFJBO0 (KubeCon + CloudNativeCon 2021)

https://thinkit.co.jp/article/19637 (後編)

上記記事で紹介されている動画よりより最近のもの
https://www.youtube.com/watch?v=sTQv4QOC-TI (KubeCon + CloudNativeCon 2022)

公式
https://gateway-api.sigs.k8s.io/

## Getting Start

### Setup

Gateway API をサポートしているプロジェクトの中のうち Istio を利用する
https://gateway-api.sigs.k8s.io/implementations/

下記の Istio ドキュメントを参考にまず試す。

https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/
https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/

```
// minikube
$ minikube -p gw-api-istio start -n 2 --driver qemu --network socket_vmnet

// Gateway API CRD
$  kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null || \
  { kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.6.1" | kubectl apply -f -; }
customresourcedefinition.apiextensions.k8s.io/gatewayclasses.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/gateways.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/httproutes.gateway.networking.k8s.io created

// helm で istio をいれる
//// ServiceAccount や Role など
$ helm install istio-base istio/base -n istio-system --create-namespace

$ helm install istiod istio/istiod -n istio-system --wait
$ kubectl get deployments -n istio-system --output wide
NAME     READY   UP-TO-DATE   AVAILABLE   AGE     CONTAINERS   IMAGES                         SELECTOR
istiod   1/1     1            1           2m18s   discovery    docker.io/istio/pilot:1.17.2   istio=pilot
```

```
// minikube tunnel を起動しておく
$ minikube -p gw-api-istio tunnel

$ kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.17/samples/httpbin/httpbin.yaml
$ kubectl get all
NAME                          READY   STATUS             RESTARTS      AGE
pod/httpbin-ff5c59f7c-xw7td   0/1     CrashLoopBackOff   6 (26s ago)   6m19s

NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/httpbin   ClusterIP   10.98.81.119   <none>        8000/TCP   6m19s

NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/httpbin   0/1     1            0           6m19s

NAME                                DESIRED   CURRENT   READY   AGE
replicaset.apps/httpbin-ff5c59f7c   1         1         0       6m19s
```

arm アーキテクチャだと起動しなかったので、自分でビルドしたイメージを代わりに利用する

```
$ docker image build -t localhost/ginoh/sample-server -f sample-server/Dockerfile sample-server/.
$ minikube -p gw-api-istio image load localhost/ginoh/sample-server
$ kubectl apply -f application.yaml
$ kubectl get all
NAME                                 READY   STATUS    RESTARTS   AGE
pod/sample-server-65b4dffb89-wfcnm   1/1     Running   0          7s

NAME                 TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/httpbin      ClusterIP   10.104.246.180   <none>        8000/TCP   7s
service/kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP    46h

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/sample-server   1/1     1            1           7s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/sample-server-65b4dffb89   1         1         1       7s
```

```
// gateway & httproute
$ kubectl create namespace istio-ingress
$ kubectl apply -f gateway.yaml

$ kubectl -n istio-ingress get gtw
NAME      CLASS   ADDRESS        PROGRAMMED   AGE
gateway   istio   10.105.39.29   True         2m49s

// istio-ingress namespace に service や deployment がデプロイされている
$ kubectl -n istio-ingress get all
NAME                                 READY   STATUS    RESTARTS   AGE
pod/gateway-istio-86f5849db4-9hjdd   1/1     Running   0          3m15s

NAME                    TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)                        AGE
service/gateway-istio   LoadBalancer   10.105.39.29   10.105.39.29   15021:32368/TCP,80:30861/TCP   3m15s

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/gateway-istio   1/1     1            1           3m15s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/gateway-istio-86f5849db4   1         1         1       3m15s

$ kubectl get httproutes
NAME   HOSTNAMES                 AGE
http   ["httpbin.example.com"]   19m

$ kubectl wait -n istio-ingress --for=condition=ready gateways.gateway.networking.k8s.io gateway
gateway.gateway.networking.k8s.io/gateway condition met

$ export INGRESS_HOST=$(kubectl get gateways.gateway.networking.k8s.io gateway -n istio-ingress -ojsonpath='{.status.addresses[*].value}')

$ curl -i -HHost:httpbin.example.com "http://$INGRESS_HOST/get"
HTTP/1.1 200 OK
date: Mon, 22 May 2023 15:46:41 GMT
content-length: 13
content-type: text/plain; charset=utf-8
x-envoy-upstream-service-time: 0
server: istio-envoy

Hello, World!

$ curl -s -HHost:httpbin.example.com "http://$INGRESS_HOST/headers"
{"Accept":["*/*"],"My-Added-Header":["added-value"],"User-Agent":["curl/7.88.1"],....}
```

デフォルトの Gateway の構成だと、Gateway のための `Deployment` および `Service` は自動でプロビジョニングされる。

自動化された Deployment を利用する場合は、

- Gateway リソースの annotation と label は `Deployment` および　`Service` にもコピーされる。
- annotation を指定することで、service の type を指定できる。デフォルトは ` LoadBalancer` だが `ClusterIP` を指定するなどができる。

設定をカスタマイズすることで手動でのプロビジョニングも可能となる。

事前にマニュアルで `Deployment`、`Service` を用意できれば以下のようにして `Service` を指定すれば `Gateway` と紐付けできる。

```
// e.g.
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: gateway
spec:
  addresses:
  - value: ingress.istio-gateways.svc.cluster.local
    type: Hostname
...
```
