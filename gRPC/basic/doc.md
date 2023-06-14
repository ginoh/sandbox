## 参考

[作ってわかる！ はじめての gRPC](https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/intro)
[Go 言語で gRPC 通信してみる（Echo サーバー＆クライアント）](https://maku.blog/p/ij4jv9k/)
スターティング gRPC (書籍)

[gRPC 公式](https://grpc.io/)
[Language Guide (proto3)](https://developers.google.com/protocol-buffers/docs/proto3)

## gRPC とは

Google が開発した RPC (Remoe Procedure Call) フレームワーク。

特徴としては、通信方式として　 HTTP/2 を利用し、データのシリアライズに Protocol Buffers を利用する。
ちなみに、gRPC の `g` の意味はバージョンによって違うらしい。

## gRPC をはじめる

作ってわかる! はじめての gRPC を進めて備忘録を残しておく。スターティング gRPC もある程度目を通しておく。

プログラミング言語としては Go を利用する。

### .proto ファイルの作成

gRPC の開発は次のような流れで開発をする。

1. スキーマ定義(関数定義など)ファイルを作成(または追加・編集)
2. スキーマ定義からボイラープレートのコードを自動生成
3. ビジネスロジックの生成

スキーマ定義ファイルは Protocol Buffers の `.proto` ファイルとして作成する。

次のような感じでスキーマ定義をする

```
// protoのバージョンの宣言
syntax = "proto3";

// protoファイルから自動生成させるGoのコードの置き先
option go_package = "pkg/grpc";

// packageの宣言
package myapp;

// サービスの定義
service GreetingService {
	// サービスが持つメソッドの定義
	rpc Hello (HelloRequest) returns (HelloResponse);
}

// 型の定義
message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string message = 1;
}
```

`syntax`:

何もしていないと `proto2` として扱われる。新しいのは `proto3` で互換性もないので `proto3` 指定しておけばいいはず。

`package`:

異なる `.proto` ファイルの型を利用するときなどに名前の衝突を防ぐために package 名を指定できる。コード生成時に与える言語ごとの影響は以下を参照するとよい。

- https://protobuf.dev/programming-guides/proto3/#packages

`service`:

リモートで呼び出す関数をメソッド、メソッドをまとめたものをサービスと呼ぶ。

ドキュメントやコードによっては、rpc メソッドの記述にセミコロンではなく `{}`を利用して以下のような記載を見かけることがある。

```
rpc Hello (HelloRequest) returns (HelloResponse) {}
```

これは option など記述できるところを単に空にしているだけで、この場合はセミコロン指定しているのと等しい。

https://stackoverflow.com/questions/30106667/grpc-protobuf-3-syntax-what-is-the-difference-between-r

`message`:

Protocol Buffers では全ての値が型を持つ。型はスカラー型とメッセージ型に大別される。メッセージ型はフィールドとしてスカラー型・メッセージ型を複数持つことができる。

スカラー型

- https://protobuf.dev/programming-guides/proto3/#scalar

メッセージ型

- https://protobuf.dev/programming-guides/proto3/#simple

メッセージ型のフィールドにはフィールド番号を割り当てる。番号はシリアライズされたフィールドの識別に使われるため、メッセージの中で一意である必要がある。使える番号や削除等の注意点は下記を参照する。

- https://protobuf.dev/programming-guides/proto3/#assigning

その他、 より多くの型については書籍やドキュメントで確認する。

### ボイラープレートコードの作成

今回は Go 言語のコードを生成するが、そのためにコンパイラと plugin をインストールする。

コンパイラのインストール

```
// Mac 用
$ brew install protobuf
```

plugin のインストール

```
$ cd sample
$ go mod init mygrpc
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

公式サイトを参考にしているが、version が新しくはなさそうだったので、`latest` を指定した。

protocol ファイルと 各種ディレクトリを記事を参考に以下で作成。

```
sample
├── api
│   └── hello.proto
├── go.mod
└── pkg
    └── grpc
```

コードの自動生成

```
$ cd api
$ protoc --go_out=../pkg/grpc --go_opt=paths=source_relative --go-grpc_out=../pkg/grpc --go-grpc_opt=paths=source_relative hello.proto
```

- `paths=source_relative` を指定することにより、入力ファイル (.proto) と同じディレクトリ構成で .go ファイルを出力する。
- `-IPATH` or `--proto_path=PATH` というオプションは　 proto ファイルの import path を指定する。`paths=source_relative`とこのオプションが指定されている場合、出力先のディレクトリ階層は `PATH` にあたるものは出力されない。フラグ指定しないとコマンドを起動したディレクトリを見るようだが、一般的には プロジェクトルートを指定しておけばよさそう。(今回は指定していない)
  - https://protobuf.dev/programming-guides/proto3/#importing

エントリポイントの実装コードをおくディレクトリとファイル(main.go)を作る。今回はビジネスロジックもここに実装する。

```
$ cd ..
$ mkdir -p cmd/server
$ touch cmd/server/main.go

$ tree .
.
├── api
│   └── hello.proto
├── cmd
│   └── server
│       └── main.go
├── go.mod
└── pkg
    └── grpc
        ├── hello.pb.go
        └── hello_grpc.pb.go
```

### gRPC サーバの実装

サーバのエントリポイント用ファイル (main.go)にサーバを起動するコードを追加する。

grpc サーバは大体以下のような感じでサーバを作成・起動すればよい

```
	listener, err := net.Listen("tcp", "8080")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()

	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()
```

これだけではハンドラが登録されていない http サーバみたいなものなので、サービスの登録をする。

`.proto` ファイルを元に自動生成されたコードに`RegisterGreetingSeriviceServer` 関数が存在する。この関数を利用することサービスを登録する。

```
RegisterGreetingServiceServer(s, [サーバーに登録するサービス])
```

第２引数もまた自動生成された interface なので、interface を満たすような構造体・メソッドを実装すればよい。

例えば、`main.go` に以下のような実装を追加する

```
import (
  ・
  ・
	hellopb "mygrpc/pkg/grpc"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
  ・
  ・
}

func NewMyServer() *myServer {
	return &myServer{}
}

func main() {
  ・
  ・
	hellopb.RegisterGreetingServiceServer(s, NewMyServer())
  ・
  ・
}
```

実装後に動作確認をする場合、http プロトコルではないため `curl` を使うことができない。そこで、`grpCurl` と呼ばれるツールを使う。

```
$ brew install grpcurl
```

また、`grpcurl` を使って動作確認するにはリフレクションの設定が必要となるので設定する。

```
import (
	・
  ・
	"google.golang.org/grpc/reflection"
)

func main() {
  ・
  ・
	reflection.Register(s)
  ・
  ・
}
```

なぜ必要になるかというと、`grpcurl` は `.proto` ファイルで定義されているシリアライズルールを把握できていないのでサーバから情報を取得する必要があるため。
https://github.com/grpc/grpc/blob/master/doc/server-reflection.md

以下のように動作確認できる。

```
$ grpcurl -plaintext localhost:8080 list
grpc.reflection.v1alpha.ServerReflection
myapp.GreetingService
$  grpcurl -plaintext -d '{"name": "ginoh"}' localhost:8080 myapp.GreetingService.Hello
{
  "message": "Hello, ginoh!"
}
```

### gRPC クラインアントの実装

エントリポイントの実装コードをおくディレクトリとファイル(main.go)を作る。今回はビジネスロジックもここに実装する。

```
$ cd ..
$ mkdir -p cmd/client
$ touch cmd/client/main.go

$ tree .
.
├── api
│   └── hello.proto
├── cmd
│   ├── client
│   │   └── main.go
│   └── server
│       └── main.go
├── go.mod
└── pkg
    └── grpc
        ├── hello.pb.go
        └── hello_grpc.pb.go
```

クライアントの実装の基本的な流れはおおよそ次の処理を記述する。

1. コネクションの確立
2. クライアントの生成 (クライアントのコード自体は自動生成)
3. クライアントのメソッドを利用してリクエスト・レスポンスを行う

コネクションの確立は `Dial()` を利用する。

```
  address := "localhost:8080"
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatal("Connection failed.")
		return
	}
	defer conn.Close()
```

この例で指定している、grpc.WithXXX に関する意味は以下。
`grpc.WithTransportCredentials(insecure.NewCredentials())`: コネクションで SSL/TLS を使用しない。

`grpc.WithBlock()`: コネクションが確立するまで待機する。

クライアントの作成はコードが自動生成されているので、確立したコネクションを利用して例えば以下のように書ける。

```
import (
  ・
  ・
	hellopb "mygrpc/pkg/grpc"
)

client = hellopb.NewGreetingServiceClient(conn)
```

あとはこのクライアントにはスキーマとして定義した Hello メソッドが実装されているため、ビジネスロジックの中で実行すればよい。

```
  // name をユーザに入力させるなどで取得してきたとする

	req := &hellopb.HelloRequest{
		Name: name,
	}
	res, err := client.Hello(context.Background(), req)
```

Server を起動した上で、Client を実行すると動作確認できる。

## ストリーミング

gRPC では 次の 4 つの通信方式がある。

- Unary RPC・・・1 リクエストに対し、1 レスポンスが返ってくる
- Server streaming RPC・・・１リクエストに対し、サーバから複数のレスポンスが返ってくる
- Client streaming RPC・・・クライアントが複数のリクエストを送り、サーバがレスポンスを 1 回返す
- Biderectional streaming RPC・・・クライアント・サーバ共に任意のタイミングでリクエスト・レスポンスを送る

https://grpc.io/docs/what-is-grpc/core-concepts/#rpc-life-cycle

ストリーミングには、サーバストリーミング・クライアントストリーミング・双方向ストリーミングがある。

### サーバストリーミングの実装

記事を参照 (TBD)

### クライアントストリーミングの実装

記事を参照 (TBD)

### 双方向ストリーミングの実装

記事を参照 (TBD)

## ステータスコード

gRPC ではメソッドの呼び出しに成功した場合、HTTP/2 上ではステータスコードは 200 OK を返却する。

メソッドの処理結果によるステータスは独自のステータスコードを利用する。これは gRPC の関心はメソッドを呼び出して結果を受け取ることであり、HTTP/2 上での実装を意識しない設計だからである。

用意されているエラーコードは以下を参照。
https://grpc.io/docs/guides/error/#error-status-codes

ちなみにこのページにはエラーハンドリングの設計についても書かれている。

gRPC のステータスコードはレスポンスヘッダ内のフィールドに格納されて送信される。

### サーバ側実装例

エラーステータスコードを返却する場合は、エラーレスポンスにスタータスコードを設定してレスポンスすればよい。

```
// e.g.
import (
  ・
  ・
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"
)

err := status.Error(codes.Unknown, "unknown error occurred")
```

ただし、もし詳細メッセージを追加したい場合は次のように `WithDetails`　メソッドを利用することで `details` フィールドを設定することができる。

```
	stat := status.New(codes.Unknown, "unknown error occurred")
	stat, _ = stat.WithDetails(<詳細メッセージ>)
	err := stat.Err()
```

<詳細メッセージ> に渡すべきメッセージ型は Protobuf 由来の構造体ならば問題ないが、gRPC では以下の `.proto` ファイルで定義されているメッセージ型を使うのが望ましい。

https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto

この `.proto` ファイルから生成された golang のコードはパッケージとして提供されているため、利用する場合はパッケージを使えばよい。

https://pkg.go.dev/google.golang.org/genproto/googleapis/rpc/errdetails

実装コードは次のようになる。

```
import (
  ・
  ・
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"
  "google.golang.org/genproto/googleapis/rpc/errdetails"
)
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
  ・
  ・
  ・
	stat := status.New(codes.Unknown, "unknown error occurred")
	stat, _ = stat.WithDetails(&errdetails.DebugInfo{
		Detail: "detail reason of err",
	})
	err := stat.Err()
	return nil, err
}
```

上記内容を `grpCurl` で動作確認すると次のような結果が得られる。

```
$ grpcurl -plaintext -d '{"name": "ginoh"}' localhost:8080 myapp.GreetingService.Hello

ERROR:
  Code: Unknown
  Message: unknown error occurred
  Details:
  1)    {"@type":"type.googleapis.com/google.rpc.DebugInfo","detail":"detail reason of err"}

```

### クライアント側実装

クライアント側では `FromError` 関数を使うことでエラーレスポンスからステータスコード・メッセージを取り出すことができる。

```
  ・
  ・
	res, err := client.Hello(context.Background(), req)
	if err != nil {
		if stat, ok := status.FromError(err); ok {
			fmt.Printf("code: %s\n", stat.Code())
			fmt.Printf("message: %s\n", stat.Message())
		} else {
			fmt.Println(err)
		}
  ・
  ・
```

詳細メッセージ の details フィールドを確認したいときはサーバ側実装と同じように `errordetails` パッケージを利用する。

ただし、何かメソッドを使うわけではなくて import して使うようだ。この init 処理が必要ということなのだろう。

https://github.com/googleapis/go-genproto/blob/e85fd2cbaebc35e54b279b5e9b1057db87dacd57/googleapis/rpc/errdetails/error_details.pb.go#L1121

```
import (
  ・
  ・
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
)

  ・
  ・
	res, err := client.Hello(context.Background(), req)
	if err != nil {
		if stat, ok := status.FromError(err); ok {
			fmt.Printf("code: %s\n", stat.Code())
			fmt.Printf("message: %s\n", stat.Message())
      fmt.Printf("details: %s\n", stat.Details())
		} else {
			fmt.Println(err)
		}
  ・
  ・

```

## インターセプタ

リクエスト・レスポンスの送受信の前後に中間処理(e.g. 認証)を挟むミドルウェアのことを gRPC ではインターセプタと呼ぶ。

Unary RPC のインターセプタは次のような形式で実装する。
https://pkg.go.dev/google.golang.org/grpc#UnaryServerInterceptor

サーバ側のインターセプタ実装例 (Unary RPC)

```
.
├── api
│   └── hello.proto
├── cmd
│   ├── client
│   │   └── main.go
│   └── server
│       ├── main.go
│       └── unaryInterceptor.go
├── go.mod
├── go.sum
└── pkg
    └── grpc
        ├── hello.pb.go
        └── hello_grpc.pb.go
```

`unaryInterceptor.go` の実装

```
func myUnaryServerInterceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("[pre] my unary server interceptor 1: ", info.FullMethod)
	res, err := handler(ctx, req)
	log.Println("[post] my unary server interceptor 1: ", res)
	return res, err
}
```

以下の部分が本来の処理部分を表している

```
res, err := handler(ctx, req)
```

この部分より前に処理を挟めば前処理になるし、実行後の `res` を本来の処理後の後処理が記述できる。

インターセプタを組み込むときは、gRPC サーバの生成処理を次のように記述する。

```
func main() {
  ・
  ・
  ・
	// Server 作成
	s := grpc.NewServer(
		grpc.UnaryInterceptor(myUnaryServerInterceptor1),
	)
	//s := grpc.NewServer()
  ・
  ・
  ・
}
```

`grpc.UnaryInterceptor` 関数を利用して `ServerOption` として `grpc.NewServer` の引数に渡す。

実行して動作確認を行う。

```
// cmd/server
$ go run ./main.go ./unaryInterceptor.go

// 別 terminalで実行
$ grpcurl -plaintext -d '{"name": "ginoh"}' localhost:8080 myapp.GreetingService.Hello
{
  "message": "Hello, ginoh!"
}

// Server 側のログ
2023/06/11 01:52:01 start gRPC server port: 8080
2023/06/11 01:53:04 [pre] my unary server interceptor 1:  /myapp.GreetingService/Hello
2023/06/11 01:53:04 [post] my unary server interceptor 1:  message:"Hello, ginoh!"

```

ここでは確認していないが、Unary と Stream ではインターセプタが異なるため上記で実装したインターセプタは Unary の通信のときにしか処理されない。

<TBD: Stream のインターセプタ>

複数のインターセプタを設定することもできる。そのときは、`grpc.UnaryInterceptor` ではなく、`grpc.ChainUnaryInterceptor` を利用する。この関数は複数のインターセプタを引数に受け取ることができる。後述する `go-grpc-middleware` でも複数のインターセプタを設定できるが、通常は `grpc.ChainUnaryInterceptor` を直接使うのがよいらしい。
https://pkg.go.dev/github.com/grpc-ecosystem/go-grpc-middleware#ChainUnaryServer

ちなみに、pkg.go.dev のドキュメントには以下のように記載されており、

> ChainUnaryInterceptor returns a ServerOption that specifies the chained interceptor for unary RPCs. The first interceptor will be the outer most, while the last interceptor will be the inner most wrapper around the real call. All unary interceptors added by this method will be chained.

呼び出しの Wrap としては最初のインターセプタがもっとも外側で Wrap していて、その後は順に元の呼び出しに近くなるように Wrap されているらしい。

javascript のイベント伝搬であるキャプチャリングとバブリングのイメージ？ 外=>中の順で前処理を行い、中=> 外の順で後処理を行うようだ。

クライアント側の実装 (Unary RPC)

```
├── api
│   └── hello.proto
├── cmd
│   ├── client
│   │   ├── main.go
│   │   └── unaryInterceptor.go
│   └── server
│       ├── main.go
│       └── unaryInterceptor.go
├── go.mod
├── go.sum
└── pkg
    └── grpc
        ├── hello.pb.go
        └── hello_grpc.pb.go
```

`unaryInterceptor.go` の実装

```
func myUnaryClientInteceptor1(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	fmt.Println("[pre] my unary client interceptor 1", method, req)
	err := invoker(ctx, method, req, res, cc, opts...)
	fmt.Println("[post] my unary client interceptor 1", res)
	return err
}
```

以下の部分が本来のリクエスト処理部分を表している。

```
err := invoker(ctx, method, req, res, cc, opts...)
```

この部分の前後に処理を記述する。

インターセプタを組み込むときはコネクション生成に利用する `Dial` 関数の引数として指定する。

```
  ・
  ・
  address := "localhost:8080"
	conn, err := grpc.Dial(
		address,
    grpc.WithUnaryInterceptor(myUnaryClientInteceptor1),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
  ・
  ・
```

実行して動作確認

```
// Server を起動しておいた状態で client を起動
$ go run ./main.go ./unaryInterceptor.go

// e.g.
[pre] my unary client interceptor 1 /myapp.GreetingService/Hello name:"ginoh"
[post] my unary client interceptor 1 message:"Hello, ginoh!"
Hello, ginoh!
```

<TBD: client で複数のインターセプタの指定>

また、go-grpc-middleware を利用するとよく使われるインターセプタを簡単に追加できる。

https://github.com/grpc-ecosystem/go-grpc-middleware

interceptor にどのようなものがあるか・どう使えばよいかなどは README.md やドキュメントサイトを確認する。

- https://github.com/grpc-ecosystem/go-grpc-middleware#interceptors
- https://github.com/grpc-ecosystem/go-grpc-middleware/tree/main/interceptors
- https://pkg.go.dev/github.com/grpc-ecosystem/go-grpc-middleware/v2@v2.0.0-rc.5/interceptors#section-directories

## メタデータの送受信

### クライアントからサーバへのメタデータ送受信

HTTP 通信の場合、付加情報を渡すのにヘッダーフィールドを利用していたが、gRPC ではメタデータを利用する。

クライアントの実装例

メタデータの付与は `context` を利用し、以下のような流れで実装する。

1. コンテキストを生成
2. `metadata.New()` でメタデータ型を生成
3. `metadata.NewOutgoingContext` で生成したメタデータをコンテキストに付与
4. クライアントのメソッド呼び出し時にコンテキストを渡す

```
// Unary の例
func Hello() {
  ・
  ・
	ctx := context.Background()
	md := metadata.New(map[string]string{"type": "unary", "from": "client"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	//res, err := client.Hello(context.Background(), req)
	res, err := client.Hello(ctx, req)
  ・
  ・
}
```

サーバの実装例

サーバ側では `metadata.FromIncomingContext` を利用することでコンテキストからメタデータを取り出すことができる。

```
// Unary の例
import (
  ・
  "google.golang.org/grpc/metadata"
  ・
)

func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Println(md)
	}
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}
```

stream の実装の場合は、コンテキストが引数としては渡ってこないので、`stream.Context()` として取得する。

動作確認

実装したサーバとクライアントを利用するとサーバ側のログで次のような出力を確認できる。

```
map[:authority:[localhost:8080] content-type:[application/grpc] from:[client] type:[unary] user-agent:[grpc-go/1.55.0]]
```

### サーバからクライアントへのメタデータ送受信

HTTP/2 上でのデータのやりとりは以下の記事で紹介されているようにフレームという単位で行っている。
https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/stream#grpc%E3%81%AE%E3%82%B9%E3%83%88%E3%83%AA%E3%83%BC%E3%83%9F%E3%83%B3%E3%82%B0%E3%82%92%E6%94%AF%E3%81%88%E3%82%8B%E6%8A%80%E8%A1%93

ヘッダーフレームのうち最初のものをヘッダー、最後のものをトレーラーと呼ぶらしい。
サーバからのメタデータ送受信はこの２つを利用する。

サーバ側の実装

```
// Unary RPC
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Println(md)
	}

	headerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "header"})
	if err := grpc.SetHeader(ctx, headerMD); err != nil {
		return nil, err
	}

	trailerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "trailer"})
	if err := grpc.SetTrailer(ctx, trailerMD); err != nil {
		return nil, err
	}
  ・
  ・
}
```

`metadata.New()` で生成したメタデータを、`SetHeader()` または `SetTrailer()` を利用して設定する。

クライアント側の実装

```
func Hello() {
  ・
  ・
  ・
	var header, trailer metadata.MD
	res, err := client.Hello(ctx, req, grpc.Header(&header), grpc.Trailer(&trailer))
  ・
  ・
		fmt.Println(header)
		fmt.Println(trailer)
		fmt.Println(res.GetMessage())
  ・
  ・
}
```

`metadata.MD` 型として宣言した変数を `grpc.Header()`、 `grpc.Trailer()` に渡した戻り値の `CallOption` をメソッドの引数とする。

動作確認

実装したサーバとクライアントを利用するとクライアント側のログで次のような出力を確認できる。

```
・
map[content-type:[application/grpc] from:[server] in:[header] type:[unary]]
map[from:[server] in:[trailer] type:[unary]]
・
```

## その他
