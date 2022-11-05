## 参考

[Azure Key Vault Provider for Secrets Store CSI Driver](https://azure.github.io/secrets-store-csi-driver-provider-azure/docs/)

## Getting Started

### Install

クラスタ構築
```
minikube -p csi-ssd-azure-provider start --driver hyperkit --insecure-registry "10.0.0.0/24"
```

CSI Driver & Azure Provider
```
helm repo add csi-secrets-store-provider-azure https://azure.github.io/secrets-store-csi-driver-provider-azure/charts
helm install csi csi-secrets-store-provider-azure/csi-secrets-store-provider-azure --namespace kube-system

kubectl -n kube-system get pods -l app=csi-secrets-store-provider-azure
```

Azure Provider の helm chart は CSI Driver の Chart の依存関係があるので一緒にインストールされる。

YAML で provider をインストールする場合、CSI Driver カスタムリソースで audience が設定されず、Workload Identity がそのままだと動作しなさそうなので Helm でインストールしたほうが楽ではある。


### Azure Key Vault にサンプル Secret を追加

以下のチュートリアルに従って作業する
https://learn.microsoft.com/ja-jp/azure/key-vault/secrets/quick-create-portal

ただし、アクセス許可モデルは `Azure ロールベースのアクセス制御` にしておく

### SecretProviderClass の作成


### Key Vault へのアクセスの設定

Azure Key Vault Provider が提供する Key Vault インスタンスへのアクセス方法は5つ

* Service Principal ** This is currently the only way to connect to Azure Key Vault from a non Azure environment.
* Pod Identity
* User-assigned Managed Identity
* System-assigned Managed Identity
* Workload Identity

#### Service Principal

Service Principal の場合は、`nodePublishSecretRef` で Secret を指定する


まず、Service Principal を 作成する

参考
* [Azure CLI で Azure サービス プリンシパルを作成する](https://learn.microsoft.com/ja-jp/cli/azure/create-an-azure-service-principal-azure-cli)
* [リソースにアクセスできる Azure AD アプリケーションとサービス プリンシパルをポータルで作成する](https://learn.microsoft.com/ja-jp/azure/active-directory/develop/howto-create-service-principal-portal?source=recommendations)
* [Azure 組み込みロール](https://learn.microsoft.com/ja-jp/azure/role-based-access-control/built-in-roles)
* [Azure CLI を使用して Azure ロールを割り当てる](https://learn.microsoft.com/ja-jp/azure/role-based-access-control/role-assignments-cli)

`az ad sp create-for-rbac` コマンドを利用して サービスプリンシパルの作成とリソースへのアクセス設定を一度に行うことが可能。

`az ad sp create-for-rbac` を実行することでアプリケーションおよびサービスプリンシパルが作成される

アプリケーションとサービスプリンシパルの関係は、[Azure Active Directory のアプリケーション オブジェクトとサービス プリンシパル オブジェクト](https://learn.microsoft.com/ja-jp/azure/active-directory/develop/app-objects-and-service-principals) を参照

```
az login

// 今回は、Key Vault Secrets User の Role とリソースグループを割り当てる例
// リソースグループは az group list で取得できる
// アクセス制御は Azure RBAC を利用。role は Key Vault Secrets User の ID を指定
az ad sp create-for-rbac \
--role  4633458b-17de-408a-b874-0445c86b69e6 \
--scopes /subscriptions/${mySubscriptionID}/resourceGroups/${myResourceGroupName} \
--name cssd-sample-app

// 実行結果で、appId、password が表示される。password はこの時に表示されるのみなので覚えておく


// list すると、"displayName": "cssd-sample-app" の アプリケーションが作成されている
az ad app list
・
・
・
```


`AZURE_CLIENT_ID` には Service Principal の `appId` を、AZURE_CLIENT_SECRET には Service Principal の `password` をいれる
```
kubectl -n cssd-azure-test create secret generic secrets-store-creds --from-literal clientid=<AZURE_CLIENT_ID> --from-literal clientsecret=<AZURE_CLIENT_SECRET>

// Label the secret
// Refer to https://secrets-store-csi-driver.sigs.k8s.io/load-tests.html for more details on why this is necessary in future releases.
kubectl -n cssd-azure-test label secret secrets-store-creds secrets-store.csi.k8s.io/used=true
```

service-principal ディレクトリにリソースファイルを置いている
```
kubectl -n cssd-azure-test apply -f sample-pod.yaml -f sample-spc-azure.yaml

kubectl -n cssd-azure-test exec busybox-secrets-store-inline -- ls /mnt/secrets-store/ 
SECRET_1

kubectl -n cssd-azure-test exec busybox-secrets-store-inline -- ls  /mnt/secrets-store/SECRET_1
0

kubectl -n cssd-azure-test exec busybox-secrets-store-inline -- cat /mnt/secrets-store/SECRET_1/0
sample-secret-value
```

SECRET_1 は alias を設定しているため。alias を設定していない場合は secret の名前になる。

0 というデータになっているのは、`objectVersionHistory` で指定した分の過去のバージョンも同期されるため (最新は0になる)


#### Pod Identity

わりとつい最近のようだが、Pod Identity は Workload Identity に置き換わるようなのでこれはもう検証しない

参考
* https://azure.github.io/secrets-store-csi-driver-provider-azure/docs/configurations/identity-access-modes/pod-identity-mode/
* https://github.com/Azure/aad-pod-identity
* https://cloudblogs.microsoft.com/opensource/2022/01/18/announcing-azure-active-directory-azure-ad-workload-identity-for-kubernetes/
* https://learn.microsoft.com/ja-jp/azure/aks/workload-identity-overview


### User-assigned/System-assigned Managed Identity

[Azure リソースのマネージド ID とは](https://learn.microsoft.com/ja-jp/azure/active-directory/managed-identities-azure-resources/overview)

[Azure Kubernetes Service でマネージド ID を使用する](https://learn.microsoft.com/ja-jp/azure/aks/use-managed-identity)

AKS ではシステム割り当てされた kubelet マネージド ID か 自身で作成したユーザ割り当てマネージド ID を利用できる。

#### AKS の作成

参考
* [クイック スタート:Azure portal を使用して Azure Kubernetes Service (AKS) クラスターをデプロイする](https://learn.microsoft.com/ja-jp/azure/aks/learn/quick-kubernetes-deploy-portal?tabs=azure-cli)
* [クイック スタート:Azure CLI を使用して Azure Kubernetes Service クラスターをデプロイする](https://learn.microsoft.com/ja-jp/azure/aks/learn/quick-kubernetes-deploy-cli)

今回は 上記ドキュメントの CLI で作成する方法を利用する。すぐに消す想定なので特にプライベートクラスタでは作らない。
AKS クラスタ作成時に独自の kubelet マネージド ID を指定しない場合は、AKS はノード リソース グループにシステム割り当て kubelet ID を作成する。

事前準備
```
// モニタリングは特に必要ないが、手順通りに aks のクラスタを作る想定なのでプロバイダ登録する
az provider register --namespace Microsoft.OperationsManagement
az provider register --namespace Microsoft.OperationalInsights
```

ドキュメントにはなかったが、aks 作成時に次のエラーがでたためProvider を登録しておく
```
Conflict({"error":{"code":"MissingSubscriptionRegistration","message":"The subscription is not registered to use namespace 'microsoft.insights'. See https://aka.ms/rps-not-found for how to register subscriptions.","details":[{"code":"MissingSubscriptionRegistration","target":"microsoft.insights","message":"The subscription is not registered to use namespace 'microsoft.insights'. See https://aka.ms/rps-not-found for how to register subscriptions."}]}})

az provider register --namespace microsoft.insights 
```

aks の作成
```
// リソースグループは作っておいた development を指定している、マネージド ID を有効化
// 既存の SSH 公開鍵を指定
az aks create -g development -n cssd-sample --enable-managed-identity --node-count 1 --enable-addons monitoring --enable-msi-auth-for-monitoring --ssh-key-value ./id_rsa.azure.pub

az aks list
az aks get-credentials --name cssd-sample -g development
kubectl get nodes
NAME                                STATUS   ROLES   AGE   VERSION
aks-nodepool1-29983327-vmss000000   Ready    agent   13m   v1.23.12
```

ask 作成時に --generate-ssh-keys を指定すれば既存の鍵を利用してくれるかと思ったがうまくいかなかった
```
az aks create -g development -n cssd-sample --enable-managed-identity --node-count 1 --enable-addons monitoring  --enable-msi-auth-for-monitoring --generate-ssh-keys
Argument '--enable-msi-auth-for-monitoring' is in preview and under development. Reference and support levels: https://aka.ms/CLI_refstatus

private key file is encrypted
```

RBAC の設定
参考
* [Azure CLI を使用して Azure ロールを割り当てる](https://learn.microsoft.com/ja-jp/azure/role-based-access-control/role-assignments-cli)
```
// 以下で コントロールプレーンと nodepool の2つ ID がでてくるので、nodepool側の ID をRBAC の assignee として設定する
az ad sp list --all --filter "servicePrincipalType eq 'ManagedIdentity'" -o table

// リソースグループで権限つけておく
az role assignment create --assignee "{assignee}" \
--role 4633458b-17de-408a-b874-0445c86b69e6 \
--resource-group development
```

補足
AKS で自動割り当てしたものは System Assigned のマネージド ID なのかと思っていたので、ドキュメント通りに以下で確認したが、どうも User Assigned に見えた (後日、また確認する)
```
az vmss list -o table
az vmss identity show -g <resource group>  -n <vmss scalset name> -o yaml
```

deploy

managed-identity ディレクトリにリソースファイルを置いている
```
kubectl -n cssd-azure-test apply -f sample-pod.yaml -f sample-spc-azure.yaml
kubectl -n cssd-azure-test exec busybox-secrets-store-inline -- ls /mnt/secrets-store/sample-secret/ 
```

必要があれば後片付けしておく (role/cluster 削除)

### Workload Identity

* [Workload Identity (Preview)](https://azure.github.io/secrets-store-csi-driver-provider-azure/docs/configurations/identity-access-modes/workload-identity-mode/)
* [Azure AD Workload Identity](https://azure.github.io/azure-workload-identity/docs/)

今回の検証の前提として、
* Service Principal を試した時の Azure AD Application RBAC が存在する
* Managed ID を試した時の AKS クラスタが残っている

Managed ID の Role割り当て を削除しておく
```
az role assignment delete --assignee "{assignee}" \
--role 4633458b-17de-408a-b874-0445c86b69e6 \
--resource-group development
```

OIDC Issuer URL の取得
```
// aks-preview extension のインストール
az extension update --name aks-preview

// AKS クラスタの OIDC Issuer の有効化
az aks update -g development -n cssd-sample --enable-oidc-issuer

// URL の取得
az aks show --resource-group development --name cssd-sample --query "oidcIssuerProfile.issuerUrl" -otsv

export SERVICE_ACCOUNT_ISSUER=<Issuer URL>
```

以下を設定
```
export APPLICATION_NAME="cssd-sample-app"
export APPLICATION_CLIENT_ID=$(az ad sp list --display-name ${APPLICATION_NAME} --query '[0].appId' -otsv)
export SERVICE_ACCOUNT_NAME="cssd-sample-sa"
export SERVICE_ACCOUNT_NAMESPACE="cssd-azure-test"
export APPLICATION_OBJECT_ID="$(az ad app show --id ${APPLICATION_CLIENT_ID} --query id -otsv)"
```

以下を設定するが、今回は、`ServiceAccount` を `cssd-sample-sa`、`namespace` を `cssd-azure-test` とする
```
export SERVICE_ACCOUNT_NAME=<name of the service account used by the application pod (pod requesting the volume mount)>
export SERVICE_ACCOUNT_NAMESPACE=<namespace of the service account>
```

federated identity credential を追加
```
cat <<EOF > params.json
{
  "name": "kubernetes-federated-credential",
  "issuer": "${SERVICE_ACCOUNT_ISSUER}",
  "subject": "system:serviceaccount:${SERVICE_ACCOUNT_NAMESPACE}:${SERVICE_ACCOUNT_NAME}",
  "description": "Kubernetes service account federated credential",
  "audiences": [
    "api://AzureADTokenExchange"
  ]
}
EOF
```
workload-identity ディレクトリにリソースファイルを置いている
```
kubectl -n cssd-azure-test create sa cssd-sample-sa
kubectl -n cssd-azure-test apply -f sample-pod.yaml -f sample-spc-azure.yaml
kubectl -n cssd-azure-test exec busybox-secrets-store-inline -- ls /mnt/secrets-store/sample-secret 
0
```

今回は学習のために手動でコマンドを実行しているが、管理しやすくするための [CLI (azwi)](https://azure.github.io/azure-workload-identity/docs/installation/azwi.html) もあるようだ


Azure AD Workload Identity の仕組みは以下のリンク先にフローが記載されている、
* https://azure.github.io/azure-workload-identity/docs/#how-it-works
* https://azure.github.io/azure-workload-identity/docs/installation/self-managed-clusters/oidc-issuer.html

端的に言うと次のように動作をしていると思われる。
* k8s 上の Workload (例えば、Pod)にSA Token(Bound Service Account Token)を Projected Volume としてマウントする
* Workload は SA Token と Azure AD のアクセストークンのリクエストを Azure AD へ送信し、トークンの交換を行う
  * SA Token は OpenID Connect のID Tokenとして利用できる
  * Azure AD は OpenID Connect の認証シーケンスに従って、ID Token の検証をする
  * Azure AD は事前に構成したフェデレーション資格情報の基づいた Token を信頼する
* Workload はアクセストークンを利用して Azure リソースへアクセスする

### Workload Identity Federation
(TBD)
https://learn.microsoft.com/ja-jp/azure/active-directory/develop/workload-identity-federation