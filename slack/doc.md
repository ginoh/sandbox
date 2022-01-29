### Motivation

Slack に システムから通知する際に、Incomming Webhook や WebAPI を利用すればできる
みたいなことは何となく知っていたが、ちゃんと仕組みをわかってなかったので試す

2021/07 現在で見たドキュメント等を参考にさせていただいている

### Preparation

目的を達成するには Slack アプリを作成することになる。アプリはワークスペースにインストールするため、
今回、個人用にワークスペースを作成することにした。(他にも利用用途があるかもしれない)

[Slack ワークスペースを作成する](https://slack.com/intl/ja-jp/help/articles/206845317-Slack-%E3%83%AF%E3%83%BC%E3%82%AF%E3%82%B9%E3%83%9A%E3%83%BC%E3%82%B9%E3%82%92%E4%BD%9C%E6%88%90%E3%81%99%E3%82%8B) 

上記を参考にしてワークスペースを作成しておく

### Slack への通知の実現方法

通知を実現するための方法として次の2種類がある
* Incomming Webhook
  * 汎用的なアプリとして用意されている
  * 発行したURLに通知内容を送信して利用する
  * 特定のチャンネル毎に webhook URLを新規に発行する

* WebAPI
  * Slack APIの一種で、システムからAPI経由でリソースを操作する
  * アプリのインストールは一度でよい
  * インタラクティブなやりとりや UIを構築可能

今回は2つとも試してみる

### Incomming Webhookによる通知

1. まず アプリを作成する

https://api.slack.com/apps にアクセスし、アプリを作成する

今回は、以下の通りに選択していく
- `From scratch` を選択
- App Name => `demo-app`、workspace は事前に作成したものを選択して作成
- Display Information で適当に Short description をいれる

2. Incomming Webhook を有効化し、新しい Webhook を追加する

アプリ画面の左ペインなどから、`Incomming Webhooks` を選択し、トグルスイッチを `On`にする  
その後、画面から `Add New Webhook to Workspace` を押下し、連携したチャンネルを選択する

連携完了すると、チャンネルにも通知がされる

3. サンプルリクエストで メッセージを送信する

Webhook 設定に記載されている Sample curl requestを使って、 Slack にメッセージが post できることを確認する
```
curl -X POST -H 'Content-type: application/json' --data '{"text":"Hello, World!"}' ${webhook URL}
```
この Webhook URL を知っていると誰でも POST できてしまうので扱いに気をつける

### WebAPIによる通知
Slack API を利用するのに Tokenを利用するが、Tokenには Bot/User の2種類の Token が存在する

それぞれの Token を利用して、チャンネルにメッセージを POST することを試す
また、今回利用する API の Reference は以下
https://api.slack.com/methods/chat.postMessage

#### Bot Token の利用

アプリの左ペインなどから `OAuth & Permissions` を選択し、Bot Token Scopes において
Reference にある必要な権限として、`chat:write` をつける。その後、アプリを Reinstall する。

メッセージを投稿したいチャンネルで、今回の `demo-app` を追加する

先ほどの Slackの設定画面の、`OAuth Tokens for Your Workspace` に `Bot User OAuth Token`
があるので、Token を利用して動作確認をする。
```
curl -X POST -H "Authorization: Bearer <token>" \
-H "Content-type: application/json" \
-d '{"channel": "development", "text": "Hello, Slack WebAPI With Bot Token"}' https://slack.com/api/chat.postMessage
```
今回はシンプルに text を送るだけ

#### User Tokenの利用
User Token を利用すると Userに成り済ました形で リソースを扱う

アプリの左ペインなどから `OAuth & Permissions` を選択し、User Token Scopes において
Reference にある必要な権限として、`chat:write` をつける。その後、アプリを Reinstall する。

メッセージを投稿したいチャンネルに自分がはいっていること

先ほどの Slackの設定画面の、`OAuth Tokens for Your Workspace` に `User OAuth Token`
があるので、Token を利用して動作確認をする。
```
curl -X POST -H "Authorization: Bearer <token>" \
-H "Content-type: application/json" \
-d '{"channel": "development", "text": "Hello, Slack WebAPI with User Token"}' https://slack.com/api/chat.postMessage
```

### References

https://christina04.hatenablog.com/entry/sending-messages-with-slack-app
=> Incomming Webhook, Web APIによる POSTの基礎


https://qiita.com/kanaxx/items/a12a523ca3143b5822b8
https://qiita.com/kanaxx/items/e74913d3db4841178533
=> Slash commandなど

https://tech.plaid.co.jp/slack_api_spec_trend_2020/
https://www.wantedly.com/companies/wantedly/post_articles/302887
=> 最新情報
