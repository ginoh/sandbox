kpromo は Artifact の promotionに関係する tool

過去存在した以下の tool を wrap したコマンド
* cip => container image promoter
* cip-mm
* cip-auditor
* gh2gcs
* krel promote-images
* promobot-files

https://github.com/kubernetes-sigs/promo-tools
### install

user
```
go install sigs.k8s.io/promo-tools/v3/cmd/kpromo@<tag>

// 今回
go install sigs.k8s.io/promo-tools/v3/cmd/kpromo@v3.3.0
```
### usage
#### Image Promotion

promoter manifest を定義して、イメージのpromoteをする

Google Artifact Registry に２つのリポジトリを作成し、一方から promotion を行う
##### 前準備

以下を参考に GARのリポジトリを2つ作るのとimageを一方のリポジトリに push しておく
[GAR Quickstart](https://cloud.google.com/artifact-registry/docs/docker/quickstart)


ServiceAccountに各リポジトリに対して権限をつける (read/write)

#### promotion の定義

promotionは以下の2つの定義を行う
* どの image を promotionするか
* promote 元レジストリと promote 先レジストリ

上記2つを一つの manifest に定義することもできるが、拡張性の観点から望ましい

参考：
* https://github.com/kubernetes-sigs/promo-tools/blob/main/docs/image-promotion.md#plain-manifest-example
* https://github.com/kubernetes-sigs/promo-tools/blob/main/docs/image-promotion.md#thin-manifests-example

promote ディレクトリ配下に thin manifest で sample 配置 (${}の部分は置き換える)

以下のコマンドを実行
```
kpromo cip --thin-manifest-dir ./promote --key-files XXXX.json --use-service-account
```
* --use-service-account を使うと、manifest の service-account を利用する
* key-files を渡していないと、activateに失敗

コマンド実行後に gcloud auth listで確認するとアカウント切り替わっているようなので、gcloud の activate 相当のことをしている？

デフォルトでは dryrunが実行され、実際の実行は `--confirm` をつける
```
kpromo cip --thin-manifest-dir ./promote --key-files <keyfile> --use-service-account --confirm
```
### Github Promotion

Githubのリリース公開されれているものを GCSに promotionする

gcloud, gsutils の利用を前提としているので、Google Cloud SDKなどをいれるとよさそう

下記のコマンドを実行する場合、

```
kpromo gh --org ginoh --repo sandbox --bucket gh-to-gcs-test-xxxx --tags gh2gcs --release-dir promo-tools   
```
* ginoh/sandbox リポジトリの gh2gcs タグのついた リリースの Assetsが対象
* gh-to-gcs-test-xxxx bucketにアップロードする
* bucket の promo-tools ディレクトリを作成し以下に保存
* コマンドを試した際はあらかじめ　project 等の設定をした上で activate しておいた
* privateのGithubリポジトリを利用するときは GITHUB_TOKEN 変数に token を設定することが可能

github からのダウンロードのみの場合は以下のコマンドを実行する
```
kpromo gh --org ginoh --repo sandbox --bucket gh-to-gcs-x
xxx --tags gh2gcs --release-dir promo-tools --download-only --output-dir ./download
```
assets に source code は入ってない