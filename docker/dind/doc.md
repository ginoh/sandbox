## dind を使う

dind の使いかたはコンテナイメージのドキュメントに大体書いてある
https://hub.docker.com/_/docker

TLS 関連の処理などは `Dockerfile`、`docker-entrypoint.sh` を見るとわかる

デフォルトではcli イメージには `DOCKER_TLS_CERTDIR` が設定されているため、
https://github.com/docker-library/docker/blob/fbbec7542dcfd052a3dd96b7085b92ee6ee064de/27/cli/Dockerfile#L161

cli をベースイメージにしている `dind` イメージでは TLS 関連のファイルが生成される
https://github.com/docker-library/docker/blob/fbbec7542dcfd052a3dd96b7085b92ee6ee064de/27/dind/dockerd-entrypoint.sh#L118

dind イメージで作成したファイルをうまくマウント設定すれば以下の `_should_tls` が真になるため、`DOCKER_HOST` 環境変数の値は `tcp://docker:2376` になる。
https://github.com/docker-library/docker/blob/fbbec7542dcfd052a3dd96b7085b92ee6ee064de/27/cli/docker-entrypoint.sh#L31

compose ファイルでは daemon が存在するコンテナの service 名を `docker` にすることでアクセスできるようになる

このファイルと同ディレクトリに存在する `compose.yaml` を使う時は以下のようにコマンドを実行する
```
$ docker compose up -d
$ docker compose exec client docker-entrypoint.sh sh
```
client のコンテナに exec する際は `docker-entrypoint.sh` を実行して TLSまわりや `DOCKER_HOST` が自動設定されるようにする