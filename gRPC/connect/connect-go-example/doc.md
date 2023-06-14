# 概要

gRPC を学習する上で以前にフューチャーアーキテクトの技術ブログで `connect` と呼ばれるプロダクトを知ったので触ってみる。

# 参考

[公式サイト](https://connect.build/docs/introduction/)

[フューチャー技術ブログ gRPC の Go 実装の新星、Connect](https://future-architect.github.io/articles/20220623a/)

あと、探せば色々最近のブログ記事とかあるはず。

# Connect とは

公式サイトにはこう書かれている

> Connect is a family of libraries for building browser and gRPC-compatible HTTP APIs: you write a short Protocol Buffer schema and implement your application logic, and Connect generates code to handle marshaling, routing, compression, and content type negotiation. It also generates an idiomatic, type-safe client in any supported language.

ブラウザと gRPC 互換の HTTP API を作成するためのライブラリらしい。既存の gRPC のコードに不満があったということで、buf と呼ばれるツール (Protobuf からのコード生成やらなんやらを行うツール)を作ったところが開発したらしい。

以前は Go 言語実装のみだったが、公式サイトには Kotlin・Swift・Web・Node.js とあってサポート言語が増えているようだ。とはいえ、今回は Go で試す。

# Getting started を試す (Go)

## ツール準備

```
$ go version
go version go1.20.4 darwin/arm64
$ mkdir connect-go-example
$ cd connect-go-example
$ go mod init example
$ go install github.com/bufbuild/buf/cmd/buf@latest
$ go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
```

以下は brew でもインストール可能

- buf
- grpcurl

```
$ brew install bufbuild/buf/buf
$ brew install grpcurl
```

`buf`, `protoc-gen-go`, `protoc-gen-connect-go` が実行できるように PATH に追加しておく。

## サービス定義

proto ファイルを作る。

```
$ mkdir -p greet/v1
$ touch greet/v1/greet.proto
```

```
syntax = "proto3";

package greet.v1;

option go_package = "example/gen/greet/v1;greetv1";

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

buf を使ってコード生成をするが、そのためには設定ファイルが必要になる。`protoc-gen-connect-go` は protoc の plugin として動作するため、protoc を使うことも可能。

一旦は buf を使ってみる。

```
$ buf mod init // 基本的な雛形ができる
$ touch buf.gen.yaml　// plugin の指定などを書いた yaml を書く
$ cat buf.gen.yaml

version: v1
plugins:
  - plugin: go
    out: gen
    opt: paths=source_relative
  - plugin: connect-go
    out: gen
    opt: paths=source_relative
```

lint および生成

```
$ buf lint
$ buf generate
$ tree .
.
├── buf.gen.yaml
├── buf.yaml
├── gen
│   └── greet
│       └── v1
│           ├── greet.pb.go
│           └── greetv1connect
│               └── greet.connect.go
├── go.mod
└── greet
    └── v1
        └── greet.proto

7 directories, 6 files
```

### protoc plugin として　 connect-go を使う場合

```
$ mkdir gen
$ protoc --go_out=gen --go_opt=paths=source_relative --connect-go_out=gen --connect-go_opt=paths=source_relative greet/v1/greet.proto

$ tree .
.
├── gen
│   └── greet
│       └── v1
│           ├── greet.pb.go
│           └── greetv1connect
│               └── greet.connect.go
└── greet
    └── v1
        └── greet.proto

7 directories, 3 files
```

protoc plugin は `--xx_out` で指定した場合、`protoc-gen-xx` のプラグインが呼ばれる。`--xx_opt` は `protoc-gen-xx` に オプションとして渡される。

## ハンドラの実装と実行

```
$ mkdir -p cmd/server
$ touch cmd/server/main.go
```

```
$ go get golang.org/x/net/http2
$ go get github.com/bufbuild/connect-go
$ go run ./cmd/server/main.go
```

## リクエストの実行

curl の場合

```
$ curl \
    --header "Content-Type: application/json" \
    --data '{"name": "Jane"}' \
    http://localhost:8080/greet.v1.GreetService/Greet
{"greeting":"Hello, ginoh!"}
```

HTTP プロトコルでもアクセスできるようだ。

grpcurl の場合

```
$ grpcurl \
    -protoset <(buf build -o -) -plaintext \
    -d '{"name": "ginoh"}' \
    localhost:8080 greet.v1.GreetService/Greet
{
  "greeting": "Hello, ginoh!"
}
```

この例では、サーバ側にリフレクションの設定をするのではなく、 Protoset file を利用するようにしている。
https://github.com/fullstorydev/grpcurl#listing-services

これは grpcurl での動作確認時には `buf` が必要になるということだろうか？

クライアントを実装してみる

```
$ mkdir -p cmd/client
$ touch cmd/client/main.go
$ tree .
├── buf.gen.yaml
├── buf.yaml
├── cmd
│   ├── client
│   │   └── main.go
│   └── server
│       └── main.go
├── gen
│   └── greet
│       └── v1
│           ├── greet.pb.go
│           └── greetv1connect
│               └── greet.connect.go
├── go.mod
├── go.sum
└── greet
    └── v1
        └── greet.proto

10 directories, 9 files
```

client の main.go

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

動作確認

```
$ go run ./cmd/client/main.go
{hsugino@hsugino-MBA-2023}% go run ./cmd/client/main.go
2023/06/11 22:42:13 Hello, ginoh!
```

## Connect Protocol の代わりに gRPC プロトコルを使う

`connect-go` は３つのプロトコルをサポートしている

- gRPC プロトコル　・・・grpc-go と互換性がある
- gRPC-Web プロトコル・・・grpc フロンエンドと中間 Proxy を使わずに相互運用できる
- Connect プロトコル・・・HTTP/1.1,HTTP/2 上で動作する HTTP ベースのシンプルなプロトコルで gRPC,gRPC-Web のいいところをとってパッケージングしている。
  - [Connect Protocol Reference](https://connect.build/docs/protocol/) にプロトコルについて記述されている。

`connect-go` はデフォルトでは ingress では３つのプロトコルをサポートしている。クライアントはデフォルトでは Connect プロトコルを使う。

gRPC, gRPC-Web プロトコルを利用するときは、クライアント生成時に `WithGRPC()`、`WithGRPCWeb` をそれぞれ利用する。

```
client := greetv1connect.NewGreetServiceClient(
  http.DefaultClient,
  "http://localhost:8080",
  connect.WithGRPC(),
)
```
