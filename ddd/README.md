Golang を利用した DDD を学習として、参考情報を元に現時点での自分の解釈を踏まえたサンプル実装

### 参考

https://zenn.dev/7oh/articles/6338a8ccd470c7
https://qiita.com/ryokky59/items/6c2b35169fb6acafce15


[ドメイン駆動設計入門 ボトムアップでわかる! ドメイン駆動設計の基本](https://www.amazon.co.jp/%E3%83%89%E3%83%A1%E3%82%A4%E3%83%B3%E9%A7%86%E5%8B%95%E8%A8%AD%E8%A8%88%E5%85%A5%E9%96%80-%E3%83%9C%E3%83%88%E3%83%A0%E3%82%A2%E3%83%83%E3%83%97%E3%81%A7%E3%82%8F%E3%81%8B%E3%82%8B%EF%BC%81%E3%83%89%E3%83%A1%E3%82%A4%E3%83%B3%E9%A7%86%E5%8B%95%E8%A8%AD%E8%A8%88%E3%81%AE%E5%9F%BA%E6%9C%AC-%E6%88%90%E7%80%AC-%E5%85%81%E5%AE%A3-ebook/dp/B082WXZVPC)

### ディレクトリ構成

```
├── Dockerfile
├── README.md
├── application
│   └── user
│       ├── data.go
│       └── user.go
├── cmd
│   └── sample-api
│       └── main.go
├── db
│   ├── README.md
│   └── migrations
│       ├── 000001_create_users_table.down.sql
│       └── 000001_create_users_table.up.sql
├── di
│   └── di.go
├── docker-compose.yaml
├── domain
│   └── model
│       └── user
│           ├── repository.go
│           └── user.go
├── go.mod
├── go.sum
├── infrastructure
│   └── mysql
│       └── user.go
├── presentation
│   └── user
│       ├── handler.go
│       └── model.go
└── utils
    └── time.go
```

現在の考え
* layard architecture を意識した構成 (domain, application, infrastructure, presentation)
 * 各ディレクトリは書籍を参考にして、意味のある単位でまとめることにした (user関連は user パッケージとする)
* db は golang-migrate/migrate を利用した Database Migrationに関係する処理

### 実行方法
TBD