### 参考

kubernetes-sigs/secrets-store-csi-driver
https://github.com/kubernetes-sigs/secrets-store-csi-driver

docs
https://secrets-store-csi-driver.sigs.k8s.io/


### Concept
https://secrets-store-csi-driver.sigs.k8s.io/concepts.html

* Pod の start/restart 時に kubelet の CSI Driver が、SecretProviderClass CR 内で指定した外部の Secret Store から secret コンテンツを取得するために、gRPC を利用して Secrets Store CSI Provider 通信する
* Pod内に tmpsfs として Volume マウントし、Secret コンテンツが volume に書き込まれる
* Pod 削除した際に、volume のクリーンアップと削除が行われる

Secrets Store CSI Driver
Secrets Store CSI Driver は daemonset で次のコンテナで構成される
* node-driver-registrar・・・ CSI driver を kubelet に登録して、CSI 呼び出しを発行する Unix ドメインソケットを認識できるようにする
* secrets-store ・・・ CSI Spec の gRPC のNode service の実装
* libness-probe ・・・ CSI driver の正常性の監視

Provider
* AWS
* Azure
* GCP
* Vault

security
* daemonset はrootで実行かつ特権 Pod である必要がある
* provider plugin は root で実行するが特権は必要ない
* シークレットを必要とするポッドのサービス アカウント トークンは、kubelet プロセスからドライバーに転送され、次にプロバイダー プラグインに転送される場合がある
* k8s 1.22 に swap memory で永続化されないように注意

CRD
* SecretProviderCalass
  * namespaced なリソース
  * Provider設定
  * Provider特有のパラメータ
* SecretProviderClassPodStatus
  * namespaced なリソース
  * SecretProviderClass と Pod のバインディング情報が含まれる
  * CSI driver が作成する
  * Pod が リソースの owner としてセットされる

SecretProviderClass は、ポッドと同じ名前空間に作成する必要がある。

### Getting Started

クラスタ準備 (今回は minikube を利用)
```
minikube -p secrets-store-csi start --driver hyperkit --insecure-registry "10.0.0.0/24"
```
profile 名が secrets-store-csi だと kube-system の pod がわかりづらくなってしまった

Secrets Store CSI Driver のインストール
```
kc -n kube-system get pods
NAME                                        READY   STATUS    RESTARTS       AGE
coredns-6d4b75cb6d-dw44j                    1/1     Running   0              105m
etcd-secrets-store-csi                      1/1     Running   0              105m
kube-apiserver-secrets-store-csi            1/1     Running   0              105m
kube-controller-manager-secrets-store-csi   1/1     Running   0              105m
kube-proxy-thkl2                            1/1     Running   0              105m
kube-scheduler-secrets-store-csi            1/1     Running   0              105m
storage-provisioner                         1/1     Running   1 (105m ago)   105m
```

helm を利用
```
helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver --namespace kube-system
```
確認
```
kubectl get po --namespace=kube-system | grep csi-secrets
csi-secrets-store-secrets-store-csi-driver-jsmgx   3/3     Running   0              80s
```

```
kubectl get crd | grep secrets-store
NAME                                                        CREATED AT
secretproviderclasses.secrets-store.csi.x-k8s.io            2022-08-20T08:51:17Z
secretproviderclasspodstatuses.secrets-store.csi.x-k8s.io   2022-08-20T08:51:17Z
```

Provider のインストール

今回は GCP Secret Manager を使うことにする
https://github.com/GoogleCloudPlatform/secrets-store-csi-driver-provider-gcp

ただし、今回は minikube を利用するので、workload identity は設定はせずに以下のドキュメントを参考にして、
`provider-adc` or `nodePublishSecretRef` を利用する
https://github.com/GoogleCloudPlatform/secrets-store-csi-driver-provider-gcp/blob/main/docs/authentication.md

```
kubectl apply -f deploy/provider-gcp-plugin.yaml

// 今回、provider-adc を利用するにあたり、provider pod に credential 設定がいるので
// yaml をとってきて変更する
// minikube の gcp-auth addon は kube-system の Pod に対しては効果がないため
wget https://raw.githubusercontent.com/GoogleCloudPlatform/secrets-store-csi-driver-provider-gcp/main/deploy/provider-gcp-plugin.yaml

namespace が kube-system => sscd-provider-gcp に変更して保存 (provider-gcp-plugin-adc.yaml)

kubectl create ns sscd-provider-gcp
kubectl apply -f provider-gcp-plugin-adc.yaml
```

サンプルの実行

`provider-adc` のための minikube 設定
https://minikube.sigs.k8s.io/docs/handbook/addons/gcp-auth/
```
// 今回は Secret Manager へのアクセス権限をもった SA を用意
export GOOGLE_APPLICATION_CREDENTIALS=<creds-path>.json
minikube -p secrets-store-csi addons enable gcp-auth
```

```
kubectl -n sscd-test apply -f sample-spc.yaml
kubectl -n sscd-test apply -f sample-app.yaml
```
mount できないと ContainerCreating のままになる。
describe、provider 側のログの確認などをする

delete に時間がかかるようになる？

memo
* 外部の Secret Store へのアクセスは provider が行うので provider が権限を持っている必要がある
  * SA をマウントして利用するなどの場合は、provider に設定する

* (おそらく) Workload Identityの場合は、Secret を利用する Pod に ServiceAccount を設定すると、Service Account Token
が Provider にフォワードされ、Provider が Secret を利用する Pod になりすまして 外部の Store にアクセスする

* pod に `nodePublishSecret` を設定すると、Secret 情報が Provider に渡され、Provider が外部の Store にアクセスする
  * `nodePublishSecret` と `provider-adc` が両方設定されていると エラーになり、Secret のマウントはされない
