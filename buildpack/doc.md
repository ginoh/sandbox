## Cloud Native Buildpacks


>transform your application source code into images that can run on any cloud.

アプリケーションソースコードを任意のクラウドで実行できるイメージに変換してくれるらしい

Cloud Native Buildpacks は [CNCFのBlog記事](https://www.cncf.io/blog/2018/10/03/cncf-to-host-cloud-native-buildpacks-in-the-sandbox/)を見た感じ CNB と略されるようだ。

ちなみに、2020/11 時点で incubation プロジェクトになった模様

## CNBを使ってみる

公式にTutorialやGuideがあるのでまずは利用してみる

### ソースコードからイメージを作成する

以下のドキュメントをまずはそのまま試す
https://buildpacks.io/docs/app-journey/
https://buildpacks.io/docs/app-developer-guide/build-an-app/

#### 事前準備
以下をインストールしておく
* Docker
* pack (CLI)

packのインストールは以下を参考にローカルPC(Mac)にいれた
https://buildpacks.io/docs/tools/pack/

```
$ brew install buildpacks/tap/pack
```
shell の Auto-Complete  ができそうだったが、bashのみっぽかったので残念

#### サンプルアプリケーションのビルドと起動

```
// サンプルリポジトリの clone
$ git clone https://github.com/buildpacks/samples

// javaのサンプルアプリ ディレクトリへ移動
$ cd samples/apps/java-maven

// イメージをビルドするようなものはありません
$ ls
./              ../             .gitattributes* .gitignore*     .mvn/           mvnw*           mvnw.cmd*       pom.xml*        src/

// 魔法のコマンドを実行 (長いけどある程度ログも貼る)
$ pack build myapp --builder cnbs/sample-builder:bionic

bionic: Pulling from cnbs/sample-builder
・
・
Digest: sha256:9bbd638d2fc9c72bd5c9eeb7a9193960b872a45a96a678b2bd73d3d331db4883
Status: Downloaded newer image for cnbs/sample-builder:bionic
bionic: Pulling from cnbs/sample-stack-run
・
・
Digest: sha256:0c11958743d20f183ee758535d4dfd8e29959c6045c50abda82a07860ea1e12d
Status: Downloaded newer image for cnbs/sample-stack-run:bionic
0.9.3: Pulling from buildpacksio/lifecycle
・
・
Digest: sha256:bc253af2edf1577717618cb3a95f0f16bb18fc9e804efbcc1b85f657d931a757
Status: Downloaded newer image for buildpacksio/lifecycle:0.9.3
・
・
===> DETECTING
[detector] samples/java-maven 0.0.1
===> ANALYZING
[analyzer] Previous image with name "myapp" not found
[analyzer] Restoring metadata for "samples/java-maven:maven_m2" from cache
===> RESTORING
[restorer] Restoring data for "samples/java-maven:maven_m2" from cache
===> BUILDING
・
・
===> EXPORTING
・
・
Successfully built image myapp
```
初回は時間がかかるが、次回からは、 cacheが効いて早くなるとのこと

```
$ docker container run -p 8080:8080 --name myapp -d myapp

// Hello, World的なのが確認できる
$ open http://localhost:8080
```
なるほど！Dockerfileなどがなくてもなぜかビルドできた。

#### Cloud Native Buildpacks について学ぶ

[Concepts](https://buildpacks.io/docs/concepts/) のドキュメントなどをみながら学ぶ

CNBにはコンポーネントとして以下が存在する
- Builder
- Buildpack
- Lifecycle
- Platform
- Stack

まず `Lifecycle` について

CNBの `Lifecycle` はアプリケーションのビルドを行い、最終的にあぷリケーションのイメージを作成するように調整するコンポーネント

Lifecycleは以下のフェーズの流れで実行される
- Detection・・・build に利用する buildpacksを見つける
- Analysys・・・build/exportで利用できるファイルをリストアする
- Build・・ソースコードからコンテナ内で実行可能な成果物に変換
- Export・・・OCIイメージの作成

上記サンプルでイメージをビルドしたときに似たようなログが確かにでていた

では `Buildpack`とは何か？

簡単にいうと以下のように記載されていた

* ソースコードを検査し、アプリケーションのビルドと実行の計画を策定する作業の単位

buildpack は最低でも次の3つのファイルから構成される
- buildpack.toml・・・buildpack のメタデータ
- bin/detect・・・buildpackを適用するかどうかの判定ロジック
- bin/build ・・・ buildpackの 実行ロジック (build スクリプト)

`Detection`フェーズとは、要は 複数の buildpack の `bin/detect`を順に実行し、利用するべき buildpack かどうかを判定する

`/bin/detect`の内容は、例えば、ソースコード中に`package.json`があるか？、 Go Srouce fileがあるか？ などなど

次に `Builder`と`Stack`について

`Stack`とは
* 簡単にいうと ビルド時のビルド環境のイメージ(build-image)と、アプリケーション実行時のベースイメージ(run-image)のこと

`Builder`とは
* アプリケーションのビルドに関する情報 (buildpack、lifecycle実装、ビルド時環境..etc)をバンドルしたイメージ
* Buildpacks、Lifecycle、Stack’s build imageで構成されている

以下の図は公式の Builderコンポーネントの図をもってきたもの
![Builder](https://buildpacks.io/docs/concepts/components/create-builder.svg "Builder")
参考：https://buildpacks.io/docs/concepts/components/builder/

図を見ると、builderイメージ内に buildpack、lifecycle、stackの実体が含まれているように見えるが、builderイメージを利用してビルドした時に lifecycle が pull されていたのは何故？ (まだ理解中..)

`Platform`は後で見る

#### サンプル以外でビルドを試す

ビルドに利用する builder の推奨ビルダーがコマンドで確認できるので確認
```
$ pack suggest-builders
```
いくつかでてくるが、今回は Google の builderを利用する

Google の Buildpacksについては以下のリポジトリや記事がある
* [Github リポジトリ](https://github.com/GoogleCloudPlatform/buildpacks)
* [Google Cloud blog (日本語訳)](https://cloud.google.com/blog/ja/products/containers-kubernetes/google-cloud-now-supports-buildpacks)

サンプルアプリを Node.js + express で適当に作ってみる
```
$ mkdir myapp & cd myapp
$ npm init // プロンプトは全部デフォルトのまま
$ npm install express -S

//以下を index.js として作る
const express = require('express')
const app = express()
const port = 3000

app.get('/', (req, res) => {
  res.send('Hello CNB!')
})

app.listen(port, () => {
  console.log(`Example app listening at http://localhost:${port}`)
})

// package.json に start scriptを追加しておく
// "start": "node index.js"
```
ビルドと実行
```
$ pack build myapp-nodejs --builder gcr.io/buildpacks/builder:v1

$ docker container run --name myapp-nodejs -p 3000:3000 -d myapp-nodejs
$ open http://localhost:3000
```
### buildpack を作成して利用する

[Create a buildpack](https://buildpacks.io/docs/buildpack-author-guide/create-buildpack/) を参考にbuildpack の作成を試す

が長いので、基本部分だけ

こんな感じの構成になるように rubyのファイル等を作成する
```
$ tree .
├── sample-app
│   ├── Gemfile
│   └── app.rb // アプリケーション
└── sample-buildpack
    ├── bin
    │   ├── build // ビルドスクリプト
    │   └── detect // 適用判定のためのスクリプト
    └── buildpack.toml // メタデータ
```
今、このとき `build/detect` は以下のような感じになっている

```
$ cat sample-buildpack/bin/detect
#!/usr/bin/env bash
set -eo pipefail

exit 1

$ cat sample-buildpack/bin/build 
#!/usr/bin/env bash
set -eo pipefail

echo "---> Ruby Buildpack"
```
pack CLIを利用してビルドする、ビルド時に buildpackを指定することが可能
```
$ pack build test-app --path ./sample-app --buildpack ./sample-buildpack --builder cnbs/sample-builder:bionic
```
=> buildpackが適用されないためエラーになる

`detect`を改良する
```
$ cat sample-buildpack/bin/detect
#!/usr/bin/env bash
set -eo pipefail

if [[ ! -f Gemfile ]]; then
   exit 100
fi
```
=> ビルドまで進む

この後、`build`をドキュメントに従って作成するとOCIイメージを作ることができる


### 参考

https://buildpacks.io/
公式

[Cloud Native BuildpackでToil減らしていこうという話](https://speakerdeck.com/jacopen/cloud-native-buildpackdetoiljian-rasiteikoutoiuhua)  2019/2 資料

[Buildpacksのビルダーをスクラッチから作ってみる](https://future-architect.github.io/articles/20201002/) 2020/10資料

[コンテナ標準化時代における次世代Buildpack『Cloud Native Buildpack』について](https://qiita.com/TakeshiMorikawa/items/c9d4eb3a866ed56a6efd) 2018/12資料


動画等

[Intro to Cloud Native Buildpacks - Terence Lee, Heroku & Emily Casey, Pivotal](https://www.youtube.com/watch?v=SK6e_ZatOaw) 2019/11

[Production CI/CD w/CNBs: Tekton, Gitlab & CircleCI(plus), Oh My! - David Freilich & Natalie Arellano](https://www.youtube.com/watch?v=RNN8XwRWGjk) 2020/12


[Why We Are Choosing Cloud Native Buildpacks at GitLab - Abubakar Siddiq, GitLab](https://www.youtube.com/watch?v=oTC-itx6ubE) 2020/9
