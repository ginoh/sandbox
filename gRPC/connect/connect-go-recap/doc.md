# 概要

Connect の Getting started の復習をする。ついで にディレクトリ構成変更、buf コマンドを使わずに　 protoc プラグインを使ってコード生成、リフレクションやヘルスチェックの設定等を試す。

# Getting started を試す (Go)

## ツール準備

```
$ go version
go version go1.20.4 darwin/arm64
$ mkdir connect-go-example
$ cd connect-go-recap
$ go mod init example
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
```

```
$ brew install bufbuild/buf/buf
$ brew install grpcurl
```

## サービス定義

proto ファイルを作る。

```
$ mkdir -p api/greet/v1
$ touch api/greet/v1/greet.proto
```

```
syntax = "proto3";

package greet.v1;

option go_package = "example/gen/greet/v1";

message GreetRequest {
  string name = 1;
}

message GreetResponse {
  string greeting = 1;
}

service GreetService {
  rpc Greet(GreetRequest) returns (GreetResponse) {}
}
```

## コード生成

```
$ mkdir pkg
$ protoc --go_out=pkg --go_opt=paths=source_relative --connect-go_out=pkg --connect-go_opt=paths=source_relative api/greet/v1/greet.proto

$ tree .
.
├── api
│   └── greet
│       └── v1
│           └── greet.proto
├── go.mod
└── pkg
    └── api
        └── greet
            └── v1
                ├── greet.pb.go
                └── v1connect
                    └── greet.connect.go
```

protoc plugin は `--xx_out` で指定した場合、`protoc-gen-xx` のプラグインが呼ばれる。`--xx_opt` は `protoc-gen-xx` に オプションとして渡される。
connect が生成するファイル名は`go_package`で指定したパッケージ名を使って、 <package name>connect という名前になるようだった。

## ハンドラの実装と実行

```
$ mkdir -p cmd/server
$ touch cmd/server/main.go
```

`cmd/server/main.go` を実装する。

```
$ go run ./cmd/server/main.go
```

## リクエストの実行

curl の場合

```
$ curl \
    --header "Content-Type: application/json" \
    --data '{"name": "ginoh"}' \
    http://localhost:8080/greet.v1.GreetService/Greet
{"greeting":"Hello, ginoh!"}
```

HTTP プロトコルでもアクセスできるようだ。

grpcurl を使う場合、proto ファイル等を指定してスキーマの情報を与えるか、リフレクション API を有効にする必要がある。 Connect を利用しつつリフレレクション API を使う場合は `https://github.com/bufbuild/connect-grpcreflect-go`　を利用する。

- https://connect.build/docs/go/grpc-compatibility/#handlers
- https://github.com/bufbuild/connect-grpcreflect-go#example

```
$ grpcurl -plaintext -d '{"name": "ginoh"}' localhost:8080 greet.v1.GreetService/Greet
{
  "greeting": "Hello, ginoh!"
}
```

この例では、サーバ側にリフレクションの設定をするのではなく、 Protoset file を利用するようにしている。
https://github.com/fullstorydev/grpcurl#listing-services

クライアントを実装してみる

```
$ mkdir -p cmd/client
$ touch cmd/client/main.go
$ tree .
├── api
│   └── greet
│       └── v1
│           └── greet.proto
├── cmd
│   ├── client
│   │   └── main.go
│   └── server
│       └── main.go
├── go.mod
├── go.sum
└── pkg
    └── api
        └── greet
            └── v1
                ├── greet.pb.go
                └── v1connect
                    └── greet.connect.go
```

`cmd/client/main.go` を実装する

```
package main

import (
    "context"
    "log"
    "net/http"

    greetv1 "example/gen/greet/v1"
    "example/gen/greet/v1/greetv1connect"

    "github.com/bufbuild/connect-go"
)

func main() {
    client := greetv1connect.NewGreetServiceClient(
        http.DefaultClient,
        "http://localhost:8080",
    )
    res, err := client.Greet(
        context.Background(),
        connect.NewRequest(&greetv1.GreetRequest{Name: "ginoh"}),
    )
    if err != nil {
        log.Println(err)
        return
    }
    log.Println(res.Msg.Greeting)
}
```

## ヘルスチェック

gRPC のヘルスチェックをどう行うべきかはドキュメントにかかれている。
https://github.com/grpc/grpc/blob/master/doc/health-checking.md

ただ、Connect では `github.com/bufbuild/connect-grpchealth-go` を利用することでヘルスチェックサポートを追加できるらしい。

- https://connect.build/docs/go/grpc-compatibility/#handlers
- https://github.com/bufbuild/connect-grpchealth-go#example

`cmd/server/main.go` を修正して単純なヘルスチェック (StaticChecker) を追加する。

動作確認

```
$ go run cmd/server/main.go

// 別ターミナル
$ grpcurl -plaintext -d '{"service": "greet.v1.GreetService"}' localhost:8080 grpc.health.v1.Health.Check
{
  "status": "SERVING"
}
```

ついでに curl の場合

```
$ curl \
-H "Content-Type: application/json" \
-d '{"service": "greet.v1.GreetService"}' \
http://localhost:8080/grpc.health.v1.Health/Check
```
