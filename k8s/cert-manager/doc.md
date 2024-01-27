## 概要

cert-manget はkubernetes のリソースとして追加された証明書・証明書発行者を利用して
証明書の取得・更新・利用のプロセスを簡素化するソフトウェア

https://cert-manager.io/

## 事前準備

```
$ minikube -p cert-manager start --driver qemu --network socket_vmnet -n 2
$ minikube -p cert-manager addons enable ingress
```



## Getting Started

NGINX ingress controller を使った例の流れを参考にする
https://cert-manager.io/docs/tutorials/acme/nginx-ingress/

実際はローカルクラスタで試すので、DNS設定しない、自己署名の証明書を利用するの条件で行う

```
$ kubectl apply -f https://raw.githubusercontent.com/cert-manager/website/master/content/docs/tutorials/acme/example/deployment.yaml


$ kubectl apply -f https://raw.githubusercontent.com/cert-manager/website/master/content/docs/tutorials/acme/example/service.yaml
```

上記URLの利用では Deployment のイメージが amd64 アーキテクチャだったのでMacで動作しない。
arm64 アーキテクチャのイメージ自体は存在するようなので、`kuard-arm64` に変更する

```
// image を変更
$ kubectl edit deployments kuard
```

Ingress の設定

```
$ kubectl apply -f https://raw.githubusercontent.com/cert-manager/website/master/content/docs/tutorials/acme/example/ingress.yaml

$ kubectl get ingress
NAME    CLASS   HOSTS             ADDRESS          PORTS     AGE
kuard   nginx   example.example.com   192.168.105.52   80, 443   3s

// http で試すと Locationでリダイレクトされてホスト名解決できないので https だけ試している
$ curl -kivL -H 'Host: example.example.com' 'https://192.168.105.52'
curl: (60) SSL certificate problem: unable to get local issuer certificate
More details here: https://curl.se/docs/sslcerts.html

curl failed to verify the legitimacy of the server and therefore could not
establish a secure connection to it. To learn more about this situation and
how to fix it, please visit the web page mentioned above.
```

cert-manager インストール

```
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml
```

自己署名証明書用のリソースを用意する

https://cert-manager.io/docs/configuration/selfsigned/

今回は自己署名証明書を使うが、k8s クラスタ内に CAを用意する方式を使う

https://cert-manager.io/docs/configuration/selfsigned/#bootstrapping-ca-issuers

```
$ kubectl apply -f ./selfsigned-ca.yaml
namespace/sandbox created
clusterissuer.cert-manager.io/selfsigned-issuer created
certificate.cert-manager.io/my-selfsigned-ca created
clusterissuer.cert-manager.io/my-ca-issuer created

$ kubectl get clusterissuer
NAME                READY   AGE
my-ca-issuer        True    26m
selfsigned-issuer   True    26m

// 今回のサンプルでは cert-manager ネームスペースの root-secret に証明書がはいっている
$ kubectl -n cert-manager get secrets
NAME                      TYPE                DATA   AGE
cert-manager-webhook-ca   Opaque              3      68m
root-secret               kubernetes.io/tls   3      25m

```
ClusterIssuer の SecretName に cert-manager ネームスペースの Secret を指定しているのが一見不思議だが、
https://cert-manager.io/docs/configuration/ca/#deployment に記述されているように Cluster Resource Namespace
に指定されているからのようだ。

この段階で試しに証明書を作ってみる
```
$ kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-selfsigned-crt
  namespace: sandbox
spec:
  subject:
    organizations:
      - MyOrg
    countries:
      - Japan
    organizationalUnits:
      - MyUnit
    localities:
      - Chiyoda-ku
    provinces:
      - Tokyo
  duration: 2160h # 90d
  dnsNames:
    - "*.example.com"
  secretName: example-selfsigned-crt
  issuerRef:
    kind: ClusterIssuer
    name: my-ca-issuer
EOF

$ kubectl -n sandbox get secrets
NAME                     TYPE                DATA   AGE
example-selfsigned-crt   kubernetes.io/tls   3      16s

$ kubectl -n sandbox get secrets example-selfsigned-crt -o yaml | yq '.data["tls.crt"]' | base64 -d > tls.crt
$ openssl x509 -in tls.crt -text -noout

$ kubectl -n sandbox delete certificate example-selfsigned-crt
```

cert-manager は ingress に annotation を書くと自動で certificate リソースを作成してくれるので annotation を付与する

```
$ kubectl annotate ingress kuard cert-manager.io/cluster-issuer=my-ca-issuer

$ kubectl get certificate
NAME                     READY   SECRET                   AGE
quickstart-example-tls   True    quickstart-example-tls   5m
```
確認してみる
```
$ curl -kivL -H 'Host: example.example.com' 'https://192.168.105.52'
```
nginx ingress controller の fake certificate が利用されている？

openssl を使って確認
```
$ openssl s_client -connect 192.168.105.52:443 -servername example.example.com </dev/null 2>/dev/null
・
・
Verification error: unable to verify the first certificate

$ kubectl get secrets quickstart-example-tls -o yaml | yq '.data["ca.crt"]' | base64 -d > ca.crt
$ openssl s_client -connect 192.168.105.52:443 -servername example.example.com -CAfile ca.crt </dev/null 2>/dev/null
・
・
Verification: OK


$ curl -kivL -H 'Host: example.example.com' 'https://192.168.105.52' --cacert ca.crt // これはダメだった
$ curl -kivL --resolve example.example.com:443:192.168.105.52 https://example.example.com --cacert ca.crt
```

resolve つかわないとダメだった。SNI (Server Name Indication) の話。


## cainjector

TBD
