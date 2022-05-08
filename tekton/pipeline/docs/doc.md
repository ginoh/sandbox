### Tektonとは

Kubernetes ベースの CI/CDのシステムを作成するための OSS フレームワーク

いくつかのプロジェクトがあるが、その中でもベースとなるプロジェクトである Pipelines は  
CI/CD のパイプラインを k8s のリソースとして定義できる機能を提供している

### Tektonのプロジェクト

Github の tektoncd OrganizationのTopレベルのリポジトリのものが Tektonのプロジェクトであり、
experimental リポジトリには Incubating プロジェクトがある

数あるプロジェクトの中で以下の4つは Core プロジェクトと呼ばれる
* Pipeline・・・CI/CD ワークフローの基礎となるビルディングブロック
* Triggers ・・・ CI/CD ワークフローの Eventトリガー
* CLI ・・・ CI/CD ワークフロー管理の CLIインタフェース
* Dashboard・・・WebUI

### Tekton Pipelines の コンセプト


以下を読むのがよい

https://tekton.dev/docs/concepts/


### Install

[Installing Tekton Pipelines](https://tekton.dev/docs/pipelines/install/)
[CLI](https://tekton.dev/docs/cli/)

```
$ minikube -p sandbox-tekton --driver hyperkit -n 2 start

// 最新
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

// Version指定
// kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.2.0/release.yaml
```

GKE の場合は Firewall の設定で 8443 (webhook の service の targetPort にあたる) を開けておく必要がある

CLIのインストール (Mac の場合)
```
brew tap tektoncd/tools
brew install tektoncd/tools/tektoncd-cli
```
brew で入れた場合、kubectl plugin として利用できるようになっている

```
kubectl plugin list                  
The following compatible plugins are available:

/usr/local/bin/kubectl-tkn
```


### Tutorial　& Example

[Getting Started](https://tekton.dev/docs/getting-started/)
[Tasks and Pipelines](https://tekton.dev/docs/pipelines/)
[tektoncd/pipeline examples](https://github.com/tektoncd/pipeline/tree/main/examples)

`Getting Started` は一瞬で終わる、`examples` は各機能の example がまとめらているので
`Tasks and Pipelines` をみながら適宜学習する


#### Taskの基本
```
$ kubectl apply -f ../sample/hello-world/task-echo-message.yaml

$ cat <<EOF | kubectl apply -f -
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: echo-message-run
spec:
  params:
    - name: message
      value: Hello, World
  taskRef:
    name: echo-message
EOF

// log
$ kubectl get taskrun
NAME               SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
echo-message-run   True        Succeeded   2m28s       2m4s

$ kubectl logs --selector=tekton.dev/taskRun=echo-message-run
Defaulted container "step-echo-1" out of: step-echo-1, step-echo-2, place-tools (init), step-init (init), place-scripts (init)
Hello, World

// CLI を使って TaskRun を自動生成する場合
$ tkn task start echo-message -p message="Hello, World"
TaskRun started: echo-message-run-lvs54

In order to track the TaskRun progress run:
tkn taskrun logs echo-message-run-lvs54 -f -n default

// TaskRunの名前は自動生成される
$ tkn taskrun logs echo-message-run-lvs54 -f -n default
[echo-1] Hello, World

[echo-2] Hello, World
```
* params でパラメータを渡すことができる
* Task は kubernetes の Podを作成する、step の一つ一つが Pod中のコンテナになる
* step は記述された順番に実行される
* step ではコンテナイメージのコマンド実行ではなく、Taskで指定した script を実行することも可能

#### Pipeline の基本

```
$ kubectl apply -f ../sample/hello-world/pipeline-hello-goodbye.yaml

$ cat <<EOF | kubectl apply -f -
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: hello-goodbye-run
spec:
  params:
    - name: hello-message
      value: Hello, Tekton
    - name: goodbye-message
      value: GoodBye, Tekton
  pipelineRef:
    name: hello-goodbye
EOF

// log
$ tkn pipelinerun logs hello-goodbye-run -f
[hello : echo-1] Hello, Tekton

[hello : echo-2] Hello, Tekton

[goodbye : echo-1] GoodBye, Tekton

[goodbye : echo-2] GoodBye, Tekton

// CLI で PipelineRun 生成も可能
$ tkn pipeline start hello-goodbye -p hello-message="Hello, Tekton!" -p goodbye-message="GoodBye, Tekton!"
PipelineRun started: hello-goodbye-run-9zcpj

In order to track the PipelineRun progress run:
tkn pipelinerun logs hello-goodbye-run-9zcpj -f -n default

$ tkn pipelinerun logs hello-goodbye-run-9zcpj -f -n default
[hello : echo-1] Hello, Tekton

[hello : echo-2] Hello, Tekton

[goodbye : echo-1] GoodBye, Tekton

[goodbye : echo-2] GoodBye, Tekton
```

#### コンテナイメージをビルドしてデプロイする

(TBD)
