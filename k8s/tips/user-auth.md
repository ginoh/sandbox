Kubernetes の認証の概要については以下のリンク先に記述されている
https://kubernetes.io/ja/docs/reference/access-authn-authz/authentication/

上記で紹介されている認証戦略を実際に試したことをメモ

## 検証環境の準備
```
minikube start -p user-auth
```

## X509クライアント証明書 の利用

クラスタの認証局に署名された有効な証明書を利用することで認証済みユーザとして判断される。そのためユーザ追加と同じ意味。

* 証明書の CommonNameがユーザ名
* 証明書の Organization がグループ

e.g.
"/CN=ginoh/O=app1/O=app2"

=> ユーザ名: ginoh, グループ: app1, app2
### 証明書の作成

`certificates.k8s.io` API を利用することで証明書のプロビジョニングができる

参考: [Manage TLS Certificates in a Cluster](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/)

鍵を作成する
```
$ openssl genrsa -out ginoh.key 4096
```
CSR を作成
```
$ openssl req -new -key ginoh.key -out ginoh.csr -subj "/CN=ginoh/O=system:authenticated"

// 確認
$ openssl req -text -noout -in ginoh.csr
```

API を利用して署名する
```
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: csr-ginoh
spec:
  signerName: kubernetes.io/kube-apiserver-client
  request: $(cat ginoh.csr | base64 | tr -d '\n')
  usages:
    - digital signature
    - key encipherment
    - client auth
EOF

certificatesigningrequest.certificates.k8s.io/csr-adol created
```
csr は base64 エンコードする  
groupsに指定されたグループは APIサーバがリソース作成時に追加する


TBD: spec.groupsについて確認する

csr 確認と承認
```
// pendingになっている
$ kubectl get csr
NAME        AGE   SIGNERNAME                            REQUESTOR       REQUESTEDDURATION   CONDITION
csr-ginoh   28s   kubernetes.io/kube-apiserver-client   minikube-user   <none>              Pending

// approve
$ kubectl certificate approve csr-ginoh
certificatesigningrequest.certificates.k8s.io/csr-ginoh approved

$ kubectl get csr                     
NAME        AGE   SIGNERNAME                            REQUESTOR       REQUESTEDDURATION   CONDITION
csr-ginoh   15m   kubernetes.io/kube-apiserver-client   minikube-user   <none>              Approved,Issued
```

REQUESTOR が minikube だからなのか、minikube-user になっていた

証明書の取得
```
$ kubectl get csr csr-ginoh -o jsonpath='{.status.certificate}' | base64 -d > ginoh.crt
```

RBACの設定
ユーザ認証はできるようになった。
次にユーザの権限 を RBACで設定する  

RBACについてはドキュメントを参照  
https://kubernetes.io/ja/docs/reference/access-authn-authz/rbac/

今回はひとまず、auth-test namespace に対して admin の権限を与える想定

namespace作成
```
$ kubectl create namespace auth-test
```

kubeconfigの準備をする

```
// masterの確認
$ kubectl cluster-info
Kubernetes control plane is running at https://127.0.0.1:53494
CoreDNS is running at https://127.0.0.1:53494/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy


// clusterの登録
$ kubectl --kubeconfig config-auth-test config set-cluster minikube-auth-test --insecure-skip-tls-verify=true --server=https://127.0.0.1:53494

// user の作成
$ kubectl --kubeconfig config-auth-test config set-credentials ginoh --client-certificate=ginoh.crt --client-key=ginoh.key --embed-certs=true

// context の作成
$ kubectl config --kubeconfig config-auth-test set-context auth-test --cluster=minikube-auth-test --user=ginoh
```

API アクセス
```
kubectl --kubeconfig config-auth-test --context auth-test -n auth-test run nginx --image nginx --port 80 
Error from server (Forbidden): pods is forbidden: User "ginoh" cannot create resource "pods" in API group "" in the namespace "auth-test"
```
この状態でアクセスしてもエラーになるので、RoleBindingの設定をする

```

$ cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ginoh-admin
  namespace: auth-test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
  - name: ginoh
    kind: User
    apiGroup: rbac.authorization.k8s.io
EOF
```

```
$ kubectl --kubeconfig config-auth-test --context auth-test -n auth-test run nginx --image nginx --port 80

pod/nginx created

$ kubectl --kubeconfig config-auth-test --context auth-test -n auth-test get pods                         
NAME    READY   STATUS    RESTARTS   AGE
nginx   1/1     Running   0          22s
```

Appendix
CertificateSigningRequest リソースは以下のドキュメントにあるように  
自動で削除される
https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/#request-signing-process

```
Approved requests: automatically deleted after 1 hour
Denied requests: automatically deleted after 1 hour
Pending requests: automatically deleted after 1 hour
```

### サービスアカウントを利用してアクセスする

ユーザアカウントではなくサービスアカウントを kubeconfig に追加してAPIアクセスすることもできる

検証用に namespace,serviceaccount を作成
```
$ kubectl create namespace auth-test2
$ kubectl -n auth-test2 create serviceaccount my-sa

// secret が自動で生成される
$ kubectl -n auth-test2 get serviceaccounts  
NAME      SECRETS   AGE
default   1         27s
my-sa     1         16s
$ kubectl -n auth-test2 get secrets
NAME                  TYPE                                  DATA   AGE
default-token-wkf7h   kubernetes.io/service-account-token   3      40s
my-sa-token-wz7k2     kubernetes.io/service-account-token   3      31s
```

auth-test2 に対して admin の権限を与える
```
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-sa-admin
  namespace: auth-test2
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
  - name: my-sa
    kind: ServiceAccount
    namespace: auth-test2
EOF
rolebinding.rbac.authorization.k8s.io/my-svc-admin created
```

my-sa 用に生成された tokenを取得する
```
$ kubectl -n auth-test2 get secrets my-sa-token-wz7k2 -o jsonpath='{.data.token}' | base64 -d > my-sa-token
```
kubeconfig に user 追加、context 作成、context切り替え
```
$ kubectl --kubeconfig config-auth-test config set-cluster minikube-auth-test --insecure-skip-tls-verify=true --server=https://127.0.0.1:53494
$ kubectl --kubeconfig config-auth-test  config set-credentials my-sa --token $(cat my-sa-token)
$ kubectl --kubeconfig config-auth-test config set-context auth-test-sa --cluster minikube-auth-test --user my-sa
```

動作確認
```
$ kubectl --kubeconfig config-auth-test --context auth-test-sa -n auth-test2 run nginx --image nginx --port 80
$ kubectl  --kubeconfig config-auth-test --context auth-test-sa -n auth-test2 get pods
```

### 参考

kubernteesの認証  
https://kubernetes.io/ja/docs/reference/access-authn-authz/authentication/

証明書APIを利用して x509証明書をプロビジョニング  
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/


