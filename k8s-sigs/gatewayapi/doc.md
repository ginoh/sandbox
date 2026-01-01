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

Gateway API は Ingress では機能が不十分のため捌けないユースケースも捌けるように進化した API のようなものらしいと聞いた気がする。
そこで詳しい背景やコンセプトを知る前にまずは試してみる。

[Getting started with Gateway API](https://gateway-api.sigs.k8s.io/guides/#installing-gateway-api) を実践してみる。

### Setup

Gateway API をサポートしているプロジェクトは以下に記載されているように複数あるが今回は Cilium を利用する。
https://gateway-api.sigs.k8s.io/implementations/

https://docs.cilium.io/en/stable/network/servicemesh/gateway-api/gateway-api/

今の stable version の v1.13.2 だと gateway 0.5.1 までのサポートのようなので、v1.14.0-snapshot.1 の Cilium を使うことにする。

```
// qemu driver だと Cilium のインストール失敗したので Docker でやる
$ minikube -p gw-api-sample start -n 2 --driver docker

// Gateway API CRD　のインストール
$ kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.6.2/standard-install.yaml
customresourcedefinition.apiextensions.k8s.io/gatewayclasses.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/gateways.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/httproutes.gateway.networking.k8s.io created
customresourcedefinition.apiextensions.k8s.io/referencegrants.gateway.networking.k8s.io created
namespace/gateway-system created
validatingwebhookconfiguration.admissionregistration.k8s.io/gateway-api-admission created
service/gateway-api-admission-server created
deployment.apps/gateway-api-admission-server created
serviceaccount/gateway-api-admission created
clusterrole.rbac.authorization.k8s.io/gateway-api-admission created
clusterrolebinding.rbac.authorization.k8s.io/gateway-api-admission created
role.rbac.authorization.k8s.io/gateway-api-admission created
rolebinding.rbac.authorization.k8s.io/gateway-api-admission created
job.batch/gateway-api-admission created
job.batch/gateway-api-admission-patch created

// cilium のインストール
$ cilium install --version v1.14.0-snapshot.1 --kube-proxy-replacement=strict --helm-set gatewayAPI.enabled=true
Flag --kube-proxy-replacement has been deprecated, This can now be overridden via `helm-set` (Helm value: `kubeProxyReplacement`).
ℹ️  Using Cilium version 1.14.0-snapshot.1
🔮 Auto-detected cluster name: gw-api-sample
🔮 Auto-detected datapath mode: tunnel
ℹ️  helm template --namespace kube-system cilium cilium/cilium --version 1.14.0-snapshot.1 --set bpf.masquerade=true,cluster.id=0,cluster.name=gw-api-sample,encryption.nodeEncryption=false,gatewayAPI.enabled=true,kubeProxyReplacement=strict,operator.replicas=1,serviceAccounts.cilium.name=cilium,serviceAccounts.operator.name=cilium-operator,tunnel=vxlan
ℹ️  Storing helm values file in kube-system/cilium-cli-helm-values Secret
🔑 Created CA in secret cilium-ca
🔑 Generating certificates for Hubble...
🚀 Creating Service accounts...
🚀 Creating Cluster roles...
🚀 Creating ConfigMap for Cilium version 1.14.0-snapshot.1...
🚀 Creating Agent DaemonSet...
🚀 Creating Operator Deployment...
⌛ Waiting for Cilium to be installed and ready...
✅ Cilium was successfully installed! Run 'cilium status' to view installation health

$ cilium status
    /¯¯\
 /¯¯\__/¯¯\    Cilium:          OK
 \__/¯¯\__/    Operator:        OK
 /¯¯\__/¯¯\    Hubble Relay:    disabled
 \__/¯¯\__/    ClusterMesh:     disabled
    \__/

DaemonSet         cilium             Desired: 2, Ready: 2/2, Available: 2/2
Deployment        cilium-operator    Desired: 1, Ready: 1/1, Available: 1/1
Containers:       cilium             Running: 2
                  cilium-operator    Running: 1
Cluster Pods:     2/2 managed by Cilium
Image versions    cilium-operator    quay.io/cilium/operator-generic:v1.14.0-snapshot.1: 1
                  cilium             quay.io/cilium/cilium:v1.14.0-snapshot.1: 2
```

Gateway API のガイドにある Sample のな中から、`Simple Gateway` を試す
https://gateway-api.sigs.k8s.io/guides/?h=crds#installing-gateway-api

[simple-gateway](https://gateway-api.sigs.k8s.io/guides/simple-gateway/) は Ingress のモデルと同じもので次の内容になっている。

- Gateway リソースと Route リソースの所有者は同じである。
- すべての HTTP トラフィックは `foo-svc` に送信される。

Gateway リソース

```
// gateway.yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: prod-web
spec:
  gatewayClassName: cilium
  listeners:
    - name: prod-web-gw
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: Same
```

今回は cilium を利用することから、`gatewayClassName` は変更している。

これは論理的な LB のインスタンス化を表現していて、その元となっているのは `cilium` という GatewayClass とのこと。アサインされた IP は `Gateway.status` に出力される。

そのほかの特徴としては、

- この LB は port 80 で 待ち受ける。
- この LB に Route リソース (どこに Traffic を流すか定義しているもの)を紐づけるが、Route リソースは同じ Namespace のもののみを許可する。

Route リソースの例

```
// httproute.yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
    - name: prod-web
  rules:
    - backendRefs:
        - name: foo-svc
          port: 8080
```

- 紐づける Gateway リソース を `parentRefs` に指定する。
- トラフィックの送信先を `backendRefs` に指定する。

`HTTPRoute` リソースは Gateway を利用する側が登録するけれど、`allowedRoutes` で指定されたところからしか紐付けを許可しない
ということなのだろう。

試しに以下の svc と nginx の deployment を用意した

```
// application.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: nginx
  name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: nginx
    spec:
      containers:
        - image: nginx
          name: nginx
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: foo-svc
spec:
  type: ClusterIP
  ports:
    - name: http
      port: 8080
      protocol: TCP
      targetPort: 80
  selector:
    app: nginx
```

Gateway と HTTPRoute を適用してみる。

```
$ kubectl create ns gateway-sample
$ kubectl -n gateway-sample apply -f gateway.yaml -f httproute.yaml
$ kubectl -n gateway-sample get gateway,httproutes
NAME                                         CLASS    ADDRESS   PROGRAMMED   AGE
gateway.gateway.networking.k8s.io/prod-web   cilium                          36s

NAME                                      HOSTNAMES   AGE
httproute.gateway.networking.k8s.io/foo               36s
```

どうもうまく動いてないようで、event を調べたところ、以下の出力があり、確かに GatewayClass が存在しない。。

```
  Conditions:
    Last Transition Time:  2023-05-14T07:06:15Z
    Message:               GatewayClass does not exist
    Observed Generation:   1
    Reason:                NoResources
    Status:                False
    Type:                  Accepted
```

helm の template のうち `.Capabilities.APIVersions.Has`の部分がうまくきいてないようだった。
https://github.com/cilium/cilium/blob/52c383566e7f22882c0e9c0f038cef110ad5c4cf/install/kubernetes/cilium/templates/cilium-gateway-api-class.yaml

原因は https://github.com/helm/helm/issues/10760 にあるように、 `helm template` は Server アクセスしないので、
Server に後から導入するような CRD の情報はもってないからということのよう。`helm install` の方は OK そうではある。

ひとまず、一旦生成されるはずの以下を手動で適用した後、gateway, httproute を適用しなおす。

```
# Source: cilium/templates/cilium-gateway-api-class.yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: cilium
spec:
  controllerName: io.cilium/gateway-controller
```

```
$ kubectl -n gateway-sample delete -f gateway.yaml -f httproute.yaml
$ kubectl -n gateway-sample apply -f gateway.yaml -f httproute.yaml
$ kubectl -n gateway-sample get gateway,httproutes
```

```
kubectl -n gateway-sample get svc
NAME                      TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
cilium-gateway-prod-web   LoadBalancer   10.100.66.241   <pending>     80:30269/TCP   2m1s
```

`type: loadBalancer` の svc が pending になっていたので、`minikube tunnel` を使う。 cilium は `type: loadBalancer` の service 作るのだろうか？

```
minikube -p gw-api-sample tunnel
Password:
✅  トンネルが無事開始しました

📌  注意: トンネルにアクセスするにはこのプロセスが存続しなければならないため、このターミナルはクローズしないでください ...

❗  cilium-gateway-prod-web service/ingress は次の公開用特権ポートを要求します:  [80]
🔑  sudo permission will be asked for it.
🏃  cilium-gateway-prod-web サービス用のトンネルを起動しています。
```

```
$ kubectl -n gateway-sample get gateway,httproutes
NAME                                         CLASS    ADDRESS     PROGRAMMED   AGE
gateway.gateway.networking.k8s.io/prod-web   cilium   127.0.0.1                64m

NAME                                      HOSTNAMES   AGE
httproute.gateway.networking.k8s.io/foo               64m

$ kubectl -n gateway-sample get svc
NAME                      TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
cilium-gateway-prod-web   LoadBalancer   10.100.66.241   127.0.0.1     80:30269/TCP   63m

$ kubectl -n gateway-sample apply -f application.yaml
$ curl -i localhost/
HTTP/1.1 200 OK
server: envoy
date: Sun, 14 May 2023 09:40:46 GMT
content-type: text/html
content-length: 615
last-modified: Tue, 28 Mar 2023 15:01:54 GMT
etag: "64230162-267"
accept-ranges: bytes
x-envoy-upstream-service-time: 1

<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

その他のサンプル

[HTTP routing](https://gateway-api.sigs.k8s.io/guides/http-routing/)

[HTTP traffic splitting](https://gateway-api.sigs.k8s.io/guides/traffic-splitting/)

[Cross-Namespace routing¶](https://gateway-api.sigs.k8s.io/guides/multiple-ns/)

## Gateway API とは何か
