## Hierarchical Namespaces

TBD (あとで概要を書く)
## Install

HNC (Hierarchical Namepsaces Controller) には 2つの Component がある
* Manager ・・・ subnamespaces の管理、policy オブジェクトの伝搬等を行うカスタムコントローラ
* kubectl plugin ・・・ ユーザが HNCの仕組みを利用したnamespace を管理するためのプラグイン

```
[memo]
Managerはsubnamespaces の管理、policy オブジェクトの伝搬、 階層構造の正当性の確認、拡張ポイント（Extension Points）の管理を行うと書いてあったが

Extension Pointsというのはi以下のこと？
https://kubernetes.io/docs/concepts/extend-kubernetes/extend-cluster/#extension-points
```



インストール方法は Githubの release ページを確認

e.g. v0.7.0  
https://github.com/kubernetes-sigs/multi-tenancy/releases/tag/hnc-v0.7.0

manager の セットアップ
```
$ HNC_VERSION=v0.7.0
$ kubectl apply -f https://github.com/kubernetes-sigs/multi-tenancy/releases/download/hnc-${HNC_VERSION}/hnc-manager.yaml

namespace/hnc-system created
Warning: apiextensions.k8s.io/v1beta1 CustomResourceDefinition is deprecated in v1.16+, unavailable in v1.22+; use apiextensions.k8s.io/v1 CustomResourceDefinition
customresourcedefinition.apiextensions.k8s.io/hierarchyconfigurations.hnc.x-k8s.io created
customresourcedefinition.apiextensions.k8s.io/hncconfigurations.hnc.x-k8s.io created
customresourcedefinition.apiextensions.k8s.io/subnamespaceanchors.hnc.x-k8s.io created
・
(snip)
・
```

plugin のセットアップ

今回は Krewを使ってインストール  
https://krew.sigs.k8s.io/

```
$ kubectl krew install hns

$ kubectl hns --version
kubectl-hns version v0.7.0
```

## HNCの基本

[QuickStart](https://github.com/kubernetes-sigs/multi-tenancy/blob/master/incubator/hnc/docs/user-guide/quickstart.md) を参考に動作確認してみる



### k8s namespace に親子関係をもたせる

以下のような組織を考えてみる
1. platform を管理する部署がある
2. 部署で管理しているplatform-xがある
3. 部署のメンバーが所属するTeam-aがある
4. Team-aには 開発(dev)と運用(ops)のメンバーがいる

上記の権限ポリシーを考える上で以下のようにリソースを作成する
* 1,2,3 => namespaceとして作成する
* 4 => Role/Rolebinding として作成する

サービスアカウントによる権限分離ではなく namespace を利用することについては、[namespaces are a much better security boundary](https://github.com/kubernetes-sigs/multi-tenancy/blob/master/incubator/hnc/docs/user-guide/concepts.md#why-ns) を参照

namespace を作成

```
$ kubectl create namespace platform-dept
$ kubectl create namespace platform-x
$ kubectl create namespace team-a
```

role/rolebinding を作成
```
$ kubectl -n team-a create role team-a-ops --verb=update --resource=deployments
$ kubectl -n team-a create role team-a-dev --verb=get --resource=deployments
$ kubectl -n team-a create rolebinding team-a-dev --role team-a-dev --user=adol
$ kubectl -n team-a create rolebinding team-a-ops --role team-a-ops --user=dogi
```

hnc の仕組みを利用して namespace に親子関係をもたせると、親から子へ RBACを伝搬させることができる

```
// hnc
$ kubectl hns set team-a --parent platform-dept
Setting the parent of team-a to platform-dept
Succesfully updated 1 property of the hierarchical configuration of team-a

$ kubectl hns set platform-x --parent team-a

$ kubectl hns tree platform-dept
platform-dept
└── team-a
    └── platform-x

// 親から子へ rolebindingが伝搬している
$ kubectl -n platform-x describe roles
Name:         team-a-dev
Labels:       app.kubernetes.io/managed-by=hnc.x-k8s.io
              hnc.x-k8s.io/inherited-from=team-a
Annotations:  <none>
PolicyRule:
  Resources         Non-Resource URLs  Resource Names  Verbs
  ---------         -----------------  --------------  -----
  deployments.apps  []                 []              [get]


Name:         team-a-ops
Labels:       app.kubernetes.io/managed-by=hnc.x-k8s.io
              hnc.x-k8s.io/inherited-from=team-a
Annotations:  <none>
PolicyRule:
  Resources         Non-Resource URLs  Resource Names  Verbs
  ---------         -----------------  --------------  -----
  deployments.apps  []                 []              [update]   
```

### subnamespaces

マルチテナント環境では、特定の namespace の admin権限はもっていても、namespace を作成する権限はないケースが多いと思われる

hnc を利用することで、権限のある namespace に`SubnamespaceAnchor`オブジェクトを生成することで 子にあたるsubnamespaceを作成できる

```
$ kubectl hns -n team-a create platform-i
$ kubectl hns -n team-a create platform-j
$ kubectl hns -n team-a create platform-k
$ kubectl hns tree team-a
team-a
├── [s] platform-i
├── [s] platform-j
├── [s] platform-k
└── platform-x

[s] indicates subnamespaces
```
subnamespaceのsubnamespaceを作成することも可能
```
$ kubectl hns -n platform-i create api
$ kubectl hns tree team-a             
team-a
├── [s] platform-i
│   └── [s] api
├── [s] platform-j
├── [s] platform-k
└── platform-x

[s] indicates subnamespaces
```
この状態で namespace を確認してみると通常の kubernetes の namespaceが作成されている
```
$ kubectl get ns         
NAME                 STATUS   AGE
api                  Active   58s
・
(snip)
・
platform-dept        Active   54m
platform-i           Active   102s
platform-j           Active   87s
platform-k           Active   85s
platform-x           Active   2m57s
team-a               Active   2m48s
```
ドキュメントにもあったが、subnamespaces の名前もユニークである必要がある、上記の例のように `api` などはやめたほうがよさそう

また、あくまで実態は namespace なので subnamespace へのリソース反映も通常の apply などと変わりないようである

subnamespace の削除は `subnamespaceanchors`(short name: subns)リソースを削除する
```
$ kubectl -n team-a delete subns platform-k
subnamespaceanchor.hnc.x-k8s.io "platform-k" deleted
```
cluster-wideな権限をもったユーザが k8s の namespaceとしてsubnamespaceを削除しようとすると以下のようにadminssion webhook により削除できない

```
$ kubectl delete namespaces platform-j
Error from server (Forbidden): admission webhook "namespaces.hnc.x-k8s.io" denied the request: The namespace platform-j is a subnamespace. Please delete the anchor from the parent namespace team-a to delete the subnamespace.
```

また、subnamespace を子としてもっている subnamespaceは削除できない (デフォルトでは)
```
$ kubectl delete subns -n team-a platform-i 
Error from server (Forbidden): admission webhook "subnamespaceanchors.hnc.x-k8s.io" denied the request: The subnamespace platform-i is not a leaf and doesn't allow cascading deletion. Please set allowCascadingDeletion flag or make it a leaf first.
```

```
// cleanup
$ kubectl delete subns -n team-a platform-j
$ kubectl delete subns -n platform-i api
```

ただし、 subnamespace としての作成ではなく、namespaceとして作成して親子関係を持たせただけのものは削除できた

namespace の削除などで階層構造が壊れてしまった場合は、[How-to](https://github.com/kubernetes-sigs/multi-tenancy/blob/master/incubator/hnc/docs/user-guide/how-to.md#use-resolve-cond)を参考にして修正する


### 親子関係の変更
hncのコントローラは RBACオブジェクトを現在の階層と同期するので、親子関係を変更するとRBACも変更される

現在の状態
```
$ kubectl hns tree platform-dept            
platform-dept
├── team-a
│   ├── [s] platform-i
│   └── platform-x
└── [s] team-b

[s] indicates subnamespaces
```
例えば、platform-x の管理を team-b に移管したような場合を考える

```
// 現在の RBAC
$ kubectl -n platform-x get rolebindings
NAME         ROLE              AGE
team-a-dev   Role/team-a-dev   12h
team-a-ops   Role/team-a-ops   12h


// 今回は team-b を subnamespace として作成
$ kubectl hns -n platform-dept create team-b
$ kubectl -n team-b create role team-b-ops --verb=update --resource=deployments
$ kubectl -n team-b create role team-b-dev --verb=get --resource=deployments
$ kubectl -n team-b create rolebinding team-b-dev --role team-a-dev --user=fina
$ kubectl -n team-b create rolebinding team-b-ops --role team-a-ops --user=reah

// 親子関係の変更
$ kubectl hns set platform-x --parent team-b
$ kubectl hns tree platform-dept            
platform-dept
├── team-a
│   └── [s] platform-i
└── [s] team-b
    └── platform-x

$ kubectl -n team-b get rolebinding
NAME         ROLE              AGE
team-b-dev   Role/team-a-dev   12h
team-b-ops   Role/team-a-ops   12h
```
ここで `platform-i` も `team-b` にうつしてみる
```
$ kubectl hns set platform-i --parent team-b
Changing the parent of platform-i from team-a to team-b

Could not update the hierarchical configuration of platform-i.
Reason: admission webhook "hierarchyconfigurations.hnc.x-k8s.io" denied the request: Illegal parent: Cannot set the parent of "platform-i" to "team-b" because it's a subnamespace of "team-a"
```
subnamespace として作成したものは親子関係の変更はできないようだ
### 異なるリソースの伝搬

デフォルトでは RBACのみが親から子の namespace に伝搬するが、設定することで RBAC以外のリソースの伝搬も可能

例えば、secret を伝搬させることができる
```
$ kubectl -n team-a create secret generic my-creds --from-literal=password=team-a-passwd

// hns で設定する
$ kubectl hns config set-resource secrets --mode Propagate

$ kubectl -n platform-x get secret
NAME                  TYPE                                  DATA   AGE
default-token-jrgtf   kubernetes.io/service-account-token   3      5h18m
my-creds              Opaque                                1      35s
```
現状、設定できる mode は `Propagate`、`Remove`、`Ignore`

* Propagate: 親から子へオブジェクトを伝搬、親で削除されたオブジェクトも子で削除される
* Remove: 存在している伝搬されたオブジェクトを全て削除
* Ignore: 削除を含むオブジェクトの伝搬を無視する

これを利用すると共通の secret といった使い方ができる

### 階層型 NetworkPolicy

NetworkPolicyを階層型にすることができる。。が
kind は デフォルトでは NetworkPolicyをサポートしていない

kind で NetworkPolicyを利用するためのヒントは以下のissue  
https://github.com/kubernetes-sigs/kind/issues/842

TBD (今後試す)

### subnamespaces deep dive

#### subnamespace の再起的な削除を行う

`allowCascadingDeletion`を設定することで可能

```
$ kubectl hns tree team-c       
team-c
└── [s] service-1
    └── [s] dev

// これはできない
$ kubectl  delete subns service-1 -n team-c

$ kubectl hns set service-1 --allowCascadingDeletion

$ kubectl  delete subns service-1 -n team-c
$ kubectl hns tree team-c
team-c
```
#### 特定の namespace へ伝搬させない

v0.7 での機能

[Limit the propagation of an object to descendant namespaces](https://github.com/kubernetes-sigs/multi-tenancy/blob/master/incubator/hnc/docs/user-guide/how-to.md#limit-the-propagation-of-an-object-to-descendant-namespaces) にあるように以下3種類の annotation をオブジェクトに指定することで可能

* propagate.hnc.x-k8s.io/select: k8s のlabelにマッチするものだけに伝搬させる
* propagate.hnc.x-k8s.io/treeSelect: namespace を指定して伝搬かそうでないかを指定
* propagate.hnc.x-k8s.io/none: `true`を設定すると伝搬されない

`none` は特定の namespaceへの伝搬をしないではなく、特定のオブジェクトを任意の namespace に伝搬しない
## HNC の仕組み
TBD (あとで追加する)
### 参考

https://kubernetes.io/blog/2020/08/14/introducing-hierarchical-namespaces/

https://github.com/kubernetes-sigs/multi-tenancy/blob/master/incubator/hnc/docs/user-guide
