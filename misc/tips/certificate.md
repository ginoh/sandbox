## 概要

検証などで TLS 通信するときに自己署名証明書等を作成する方法について記載しておく

### openssl を使う

https://qiita.com/sanyamarseille/items/46fc6ff5a0aca12e1946
https://kubernetes.io/ja/docs/tasks/administer-cluster/certificates/


シンプルな手順だけ
```
// 秘密鍵作成
$ openssl genrsa -out server.key 2048

// CSR 作成, 質問に答える形で必要な情報を入力
$ openssl req -new -key server.key -out server.csr
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:JP
State or Province Name (full name) [Some-State]:Tokyo
Locality Name (eg, city) []:Chiyoda-ku
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Company
Organizational Unit Name (eg, section) []:Section
Common Name (e.g. server FQDN or YOUR name) []:ginoh.example.com
Email Address []:

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:
An optional company name []:


// 別のやり方で config ファイルをあらかじめ用意しておいて指定する形でもよい
// e.g. csr.conf
// $ openssl req -new -key server.key -out server.csr -config csr.conf
// https://kubernetes.io/ja/docs/tasks/administer-cluster/certificates/#openssl の例にあるような形


// CSR の中身確認
$ openssl req -in server.csr -text -noout
・
・
// 証明書の作成
$ openssl x509 -req -days 365 -signkey server.key -in server.csr -out server.crt
Certificate request self-signature ok
subject=C=JP, ST=Tokyo, L=Chiyoda-ku, O=Company, OU=Section, CN=ginoh.example.com

$ openssl x509 -in server.crt -text -noout
・
・
```
認証局 も作る場合は、https://kubernetes.io/ja/docs/tasks/administer-cluster/certificates/#openssl を参考にする



### mkcert を使う

開発用の証明書を作成するツールとして [mkcert](https://github.com/FiloSottile/mkcert) がある

```
$ brew install mkcert
```

mkcert を使って証明書をつくっていく、仕組み的には以下の記事で説明されている
https://qiita.com/k_kind/items/b87777efa3d29dcc4467

```
// 認証局をインストールしてくれる
// mkcert -CAROOT で表示されるパスにファイルが保存される
// Mac の場合、キーチェーンの証明書リストを確認すると増えていることがわかる
$ mkcert -install
Created a new local CA 💥
Sudo password:
The local CA is now installed in the system trust store! ⚡️

$ ls "$(mkcert -CAROOT)"
rootCA-key.pem  rootCA.pem

$ mkcert example.com
Created a new certificate valid for the following names 📜
 - "example.com"

The certificate is at "./example.com.pem" and the key at "./example.com-key.pem" ✅

It will expire on 24 April 2026 🗓

$ ls
example.com-key.pem     example.com.pem

// 確認
$ openssl x509 -in example.com.pem -text -noout
・
・
```
Subject に設定される値などは以下を参考にするとわかる
https://github.com/FiloSottile/mkcert/blob/2a46726cebac0ff4e1f133d90b4e4c42f1edf44a/cert.go#L50

### cert-manager

https://zenn.dev/t_ume/articles/9407eed5c64a10

```
$ kubectl create ns test-self-signed
$ cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: self-signed-sample
  namespace: test-self-signed
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: self-signed-sample
  namespace: test-self-signed
spec:
  subject:
    organizations:
      - MyOrg
    countries:
      - Japan
    localities:
      - Chiyoda-ku
    provinces:
      - Tokyo
  dnsNames:
    - "*.example.com"
  secretName: self-signed-example-com
  issuerRef:
    name: self-signed-sample
    kind: Issuer
EOF
```

```
$ kubectl -n test-self-signed get secrets self-signede-example-com -o yaml | yq '.data["tls.crt"]' | base64 -d > tls.crt
$ kubectl -n test-self-signed get secrets self-signede-ses -o yaml | yq '.data["tls.key"]' | base64 -d > tls.key
```

確認
```
$ openssl x509 -in tls.crt -text -noout
```

実際は最初に自己署名証明書を作ってCA をつくり、CAで署名した証明書を使うとよさそう



### k8s certificate API

https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/