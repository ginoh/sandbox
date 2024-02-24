
## 概要


## 参考

https://pipecd.dev/


## QuickStart (Mac で頑張って諦めた記録)

// k8s クラスタ構築
```
$ minikube -p pipecd-quickstart start --driver qemu --network socket_vmnet -n 2
```

CLI ダウンロード

リポジトリを見る限り arm64 版もあるようなので arm64版をダウンロード
```
$ curl -Lo ./pipectl https://github.com/pipe-cd/pipecd/releases/download/v0.46.0/pipectl_v0.46.0_darwin_arm64

$ chmod a+x ./pipectl
$ pipectl help

// 個人的に path を通しているところにおく
$ mv ~/bin
```

なんとなく、コマンドの出力を見て、`cobra` で作っているのかなと感じることが結構ある


quickstart を実行する。
```
pipectl quickstart --version v0.46.0
Installing the controlplane in quickstart mode...
Release "pipecd" does not exist. Installing it now.
NAME: pipecd
LAST DEPLOYED: Sat Feb 24 18:52:22 2024
NAMESPACE: pipecd
STATUS: deployed
REVISION: 1
TEST SUITE: None

Intalled the controlplane successfully!

```
このまま止まって時間がたつと失敗する

Pod がうまく起動していない
```
$ kubectl -n pipecd get pods  
NAME                              READY   STATUS             RESTARTS      AGE
pipecd-cache-56c7c65ddc-d8lgz     1/1     Running            0             102s
pipecd-gateway-58589b55f9-f9fgd   0/1     CrashLoopBackOff   3 (14s ago)   102s
pipecd-minio-677999d5bb-hf88r     0/1     CrashLoopBackOff   3 (49s ago)   102s
pipecd-mysql-6db6688bd4-gggbb     1/1     Running            0             102s
pipecd-ops-69d495dbf9-qc4zj       0/1     Init:0/1           0             102s
pipecd-server-7db9fcf8cb-wqrcf    0/1     Init:0/1           0             102s
```

`CrashLoopBackOff` に関連してまず見てみると、おそらく arm64 対応のイメージではないのだろう
```
$ kubectl -n pipecd logs pipecd-gateway-58589b55f9-f9fgd  
exec /usr/local/bin/envoy: exec format error

$ kubectl -n pipecd logs pipecd-minio-677999d5bb-hf88r   
exec /usr/bin/docker-entrypoint.sh: exec format error
```

deployment リソースを編集してみる
```
// envoy-alpine に arm64版がない
$ kubectl -n pipecd edit deploy pipecd-gateway

// minio/minio:RELEASE.2020-08-26T00-00-49Z にはarm64版がなさそう
$ kubectl -n pipecd edit deploy pipecd-minio
```
それぞれ利用イメージを envoy-alpine => `envoy:v1.21-latest`, minio => `minio:latest` にしてとりあえずやってみる
```
$ kubectl -n pipecd get pods                       
NAME                              READY   STATUS             RESTARTS      AGE
pipecd-cache-56c7c65ddc-d8lgz     1/1     Running            0             34m
pipecd-gateway-678495f867-mm6jt   1/1     Running            0             4m8s
pipecd-minio-8564c997d9-r6j47     1/1     Running            0             17m
pipecd-mysql-6db6688bd4-gggbb     1/1     Running            0             34m
pipecd-ops-69d495dbf9-qc4zj       0/1     CrashLoopBackOff   8 (74s ago)   34m
pipecd-server-7db9fcf8cb-wqrcf    0/1     CrashLoopBackOff   8 (81s ago)   34m
```
envoy は自分の環境だとこのバージョン以上にあげると起動しなかった

pipecd-ops/pipecd-server が `CrashLoopBackOff` に変わった、ログを見るとどうも同じ理由に見える
コンテナイメージとしては、`ghcr.io/pipe-cd/pipecd:v0.46.0` が利用されていそう

見た限りでは arm64版はありそうなのだが
https://github.com/pipe-cd/pipecd/pkgs/container/pipecd/175292302?tag=v0.46.0

ひとまず digest 指定 `ghcr.io/pipe-cd/pipecd@sha256:a209a01a886d02db582172a089f3e17ca36154399d900df72b711ce3a8c34c55` に書き換える

=> 結果が変わらない。。。、仕方ないのでローカルでビルドしてみる

```
git clone git@github.com:pipe-cd/pipecd.git
```
`cmd/pipecd/Dockerfile` がビルドできるといいような気がするが、artifacts が必要そうなので Make して作る

ただし、そのままビルドするとコンテナイメージにした時に動かないので、architectureを指定する
```
$ BUILD_OS=linux BUILD_ARCH=arm64 make build
$ docker image build -f cmd/pipecd/Dockerfile -t localhost/pipecd:20240224 .
```
加えて、go と yarn が利用できる環境が必要

minikube に image load する
```
$ minikube -p pipecd-quickstart image load localhost/pipecd:20240224
```
なんかイメージのロードに失敗する。。。どうも↓のように思える
https://github.com/kubernetes/minikube/issues/18021

workaround がのっていたので実行し、改めてマニフェストを書き直す
```
$ docker image save -o image.tar localhost/pipecd:20240224
$ minikube -p pipecd-quickstart image load image.tar

$ kubectl -n pipecd edit deploy pipecd-ops
$ kubectl -n pipecd edit deploy pipecd-server
```
全てのPodがたちあがるようになった
```
$ kubectl -n pipecd get pods
NAME                              READY   STATUS    RESTARTS   AGE
pipecd-cache-56c7c65ddc-d8lgz     1/1     Running   0          3h4m
pipecd-gateway-678495f867-mm6jt   1/1     Running   0          154m
pipecd-minio-8564c997d9-r6j47     1/1     Running   0          167m
pipecd-mysql-6db6688bd4-gggbb     1/1     Running   0          3h4m
pipecd-ops-58fb76cc8b-h82lc       1/1     Running   0          3s
pipecd-server-cb77b44c6-99qss     1/1     Running   0          103s
```

WebUIにアクセス
```
$ minikube -p pipecd-quickstart service -n pipecd pipecd
```

うまくいっているように思っていたが、どうも `quickstart` コマンドはインストールからいくつか実行される処理
があるようなので、仕方ないので、pipectl のインストールを一時的にコメントアウトしたものを作る

このあたりをコメントアウトしてビルドしてもう一回実行する
https://github.com/pipe-cd/pipecd/blob/master/pkg/app/pipectl/cmd/quickstart/quickstart.go#L114

ドキュメントにある通りのリダイレクトとターミナルに以下が表示されたのでドキュメント通りにやる。
```
Installing the piped for quickstart...

Installing the piped for quickstart...

Openning PipeCD control plane at http://localhost:8080/
Please login using the following account:
- Username: hello-pipecd
- Password: hello-pipecd
For more information refer to https://pipecd.dev/docs/quickstart/

Fill up your registered Piped information:
ID: XXXXXXXX
Key: YYYYYYY
✔ GitRemoteRepo: https://github.com/ginoh/examples.git█
Failed to install piped!!
github.com/pipe-cd/pipecd/pkg/app/pipectl/cmd/quickstart.(*command).run
        /Users/hsugino/ghq/github.com/ginoh/ginoh/sandbox/pipecd/pipecd/pkg/app/pipectl/cmd/quickstart/quickstart.go:125
・
・
2024/02/24 23:14:43 exit status 1: Error: ghcr.io/pipe-cd/chart/piped:v0.46.0-1-g4285674-dirty: not found
```
エラーになってしまった。ローカルでビルドしているからバージョンが特殊なものになっているのだろう。
この後の処理である pipedインストール処理でも似たようなことが起きそう。

ここまでくると大変なので、一旦 Linux用のコンテナイメージ使う方がよさそうなので、GKE使って試すことにする。

## QuickStart　(GKE使う)