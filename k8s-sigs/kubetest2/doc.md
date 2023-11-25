## 概要

kubernetes クラスタをデプロイし、クラスタに対し ent-to-end テストを実行するためのフレームワーク。
kubetest の次期版らしく、kubetest2 という名前のようだ。


kubetest2は3つの役割を持つ実行可能ファイルで構成されている。
* kubetest2 ・・・ deplyer と testerの検出と実行
* kubetest2-<DEPLOYER>・・・ kubernetesク ラスタのライフサイクル管理
* kubetest2-tester-<TESTER> ・・・ kubernetes クラスタのテスト


設計として以下を意図している。
* deployer と tester の連携を最小限にする
* 新しい deployer/tester は out-of-tree で実装することを推奨する
* kubetet2 の依存関係/表面積を小さく保つ


## Installation

```
// kubetest2 と全ての deployer/tester のインストール
go install sigs.k8s.io/kubetest2/...@latest

// 特定の deployer をインストール
go install sigs.k8s.io/kubetest2/kubetest2-DEPLOYER@latest

// 特定の tester をインストール
go install sigs.k8s.io/kubetest2/kubetest2-tester-TESTER@lates
```

## Hello, kubetest2


kubetest2 はリファレンス実装として以下がある。

deployter
* gce
* gke
* kind
* noop

tester
* clusterloader2
* exec
* ginkgo
* node


今回は簡単に試せる `kind` deployer と `exec` tester を試してみる。

事前に kind をインストールしておく。

```
$ brew install kind
```

kind deployer でできることはヘルプで見える。

```
$ kubetest2 kind --help
Usage:
  kubetest2 kind [Flags] [DeployerFlags] -- [TesterArgs]
・
・
・
DeployerFlags(kind):
      --alsologtostderr                  log to standard error as well as files
      --build-type string                --type for kind build node-image
      --cluster-name string              the kind cluster --name
      --config string                    --config for kind create cluster
      --image-name string                the image name to use for build and up
      --kube-root string                 --kube-root for kind build node-image
      --kubeconfig string                --kubeconfig flag for kind create cluster
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```


kind クラスタを作成する
```
$ kubetest2 kind --cluster-name kt2-cluster --up
I1115 08:00:27.761451   90714 app.go:61] The files in RunDir shall not be part of Artifacts
I1115 08:00:27.761466   90714 app.go:62] pass rundir-in-artifacts flag True for RunDir to be part of Artifacts
I1115 08:00:27.761476   90714 app.go:64] RunDir for this run: "/Users/hsugino/ghq/github.com/ginoh/ginoh/sandbox/k8s-sigs/kubetest2/_rundir/7b013676-de3a-4b98-9d9f-ca5fa3431855"
I1115 08:00:27.769520   90714 app.go:130] ID for this run: "7b013676-de3a-4b98-9d9f-ca5fa3431855"
I1115 08:00:27.769542   90714 up.go:63] Up(): creating kind cluster...
Creating cluster "kt2-cluster" ...
 • Ensuring node image (kindest/node:v1.27.3) 🖼  ...
 ✓ Ensuring node image (kindest/node:v1.27.3) 🖼
 • Preparing nodes 📦   ...
 ✓ Preparing nodes 📦 
 • Writing configuration 📜  ...
 ✓ Writing configuration 📜
 • Starting control-plane 🕹️  ...
 ✓ Starting control-plane 🕹️
 • Installing CNI 🔌  ...
 ✓ Installing CNI 🔌
 • Installing StorageClass 💾  ...
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kt2-cluster"
You can now use your cluster with:

kubectl cluster-info --context kind-kt2-cluster

Thanks for using kind! 😊
```

`--down` 　フラグをつけると作成後にクラスタを削除する。t

単に `--up` と `--down` だけの指定だと作成後にすぐ削除されるだけになるため、`--test` フラグを利用することで
作成 =>  テスト => 削除　の流れになる。

また、`--build` フラグを利用すると k8s をソースコードからビルドするようだ。

ソースコードがローカルで見つからない場合は以下のエラーがでる
```
・・・
I1114 22:52:08.531811   89845 build.go:46] Build(): building kind node image...
ERROR: error building node image: error finding kuberoot: could not find Kubernetes source under current working directory or GOPATH=
・・・
```

exec tester でできることを help で確認する
```
kubetest2 kind --test=exec --help
Usage:
  kubetest2 kind [Flags] [DeployerFlags] -- [TesterArgs]
・
・
<kind deployerのフラグ>
・
・
TesterArgs(exec):
kubetest2 --test=exec --  [TestCommand] [TestArgs]
  TestCommand: the command to invoke for testing
  TestArgs:    arguments passed to test command
```

クラスタ作成して、`kubectl get ns`でテストした後に、クラスタ削除する。
(kubectl がパスにあるものとする)
```
$ kubetest2 kind --up --down --test=exec -- kubectl get ns
I1125 09:17:22.884956   40803 app.go:61] The files in RunDir shall not be part of Artifacts
I1125 09:17:22.884978   40803 app.go:62] pass rundir-in-artifacts flag True for RunDir to be part of Artifacts
I1125 09:17:22.885006   40803 app.go:64] RunDir for this run: "/Users/hsugino/bin/_rundir/9621b92f-85a1-4925-a392-686c76d041f6"
I1125 09:17:22.891325   40803 app.go:130] ID for this run: "9621b92f-85a1-4925-a392-686c76d041f6"
I1125 09:17:22.891372   40803 up.go:63] Up(): creating kind cluster...
Creating cluster "kind" ...
 • Ensuring node image (kindest/node:v1.27.3) 🖼  ...
 ✓ Ensuring node image (kindest/node:v1.27.3) 🖼
 • Preparing nodes 📦   ...
 ✓ Preparing nodes 📦 
 • Writing configuration 📜  ...
 ✓ Writing configuration 📜
 • Starting control-plane 🕹️  ...
 ✓ Starting control-plane 🕹️
 • Installing CNI 🔌  ...
 ✓ Installing CNI 🔌
 • Installing StorageClass 💾  ...
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Thanks for using kind! 😊
NAME                 STATUS   AGE
default              Active   5s
kube-node-lease      Active   5s
kube-public          Active   5s
kube-system          Active   5s
local-path-storage   Active   1s
I1125 09:17:37.100998   40803 down.go:33] Down(): deleting kind cluster...
Deleting cluster "" ...
Deleted nodes: ["kind-control-plane"]
```


## deplyer/tester の実装

TBD

命名規則に沿ったファイル名をつけて、`PATH` に配置すると deplyer と tester を検出して利用可能になる。