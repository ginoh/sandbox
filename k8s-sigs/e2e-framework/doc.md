## 　概要

kubernetes クラスタで実行中のコンポーネントの end-to-end テストのため Go フレームワーク。

Go のネイティブテストAPIで end-to-end テストスイートを提供することを目的としている。
つまり、`go test`　の実行で e2e テストを実行できるフレームワークを提供する。

リポジトリ：https://github.com/kubernetes-sigs/e2e-framework

## Hello, e2e-framework

kind クラスタを作成し、Deployment リソースを登録し、正常にPodが作成されていることを確認することをテストする。

リポジトリには example コードが結構あるので参考にする
https://github.com/kubernetes-sigs/e2e-framework/tree/4626c498dd24fec202b4f7855da01efbbb78d9a0/examples


```
$ go mod init github.com/ginoh/e2e-sample
$ mkdir sample
$ touch sample/main_test.go sample/deployment_test.go

// main_test.goを記述するのに go get しておく
$ go get -v sigs.k8s.io/e2e-framework/pkg/env
$ go get -v sigs.k8s.io/e2e-framework/pkg/envconf
$ go get -v sigs.k8s.io/e2e-framework/pkg/envfuncs
$ go get -v sigs.k8s.io/e2e-framework/support/kind
```

main_test.go に TestMain関数を書く
```
func TestMain(m *testing.M) {
	//cfg, _ := envconf.NewFromFlags()
	//testEnv = env.NewWithConfig(cfg)
	testEnv, _ = env.NewFromFlags()
	kindClusterName := envconf.RandomName("sample-cluster", 16)
	namespace := envconf.RandomName("sample-ns", 16)

	// Use pre-defined environment funcs to create a kind cluster prior to test run
	testEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.CreateNamespace(namespace),
	)

	// Use pre-defined environment funcs to teardown kind cluster after tests
	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(kindClusterName),
	)

	// launch package tests
	os.Exit(testEnv.Run(m))
}
```

deploymentの登録・検証をするテストとして、`deployment_test.go` に書く。

主に features を定義した上で `Environment` 構造体のテストを呼べばいい。
```
	deploymentFeature := features.New("appsv1/deployment").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
            ・
            ・
		}).
		Assess("deployment creation", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
            ・
            ・
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
            ・
            ・
		}).Feature()

	testEnv.Test(t, deploymentFeature)
```

テストを実行する

テスト実行時に利用される `kind` はあらかじめインストールしてあってもいいが、ない場合は、`go install` を利用して自動でインストールされる。
```
$ go test -v ./sample
=== RUN   TestDeployment
=== RUN   TestDeployment/appsv1/deployment
=== RUN   TestDeployment/appsv1/deployment/deployment_creation
    deployment_test.go:49: deployment found: test-deployment
--- PASS: TestDeployment (15.03s)
    --- PASS: TestDeployment/appsv1/deployment (15.03s)
        --- PASS: TestDeployment/appsv1/deployment/deployment_creation (0.01s)
PASS
ok      github.com/ginoh/e2e-sample/sample      65.255s


$ which kind
~/go/1.21.4/bin/kind
```

## コンテナイメージ内でテスト

実際の運用の中でテスト実行していくことを考えるとコンテナ内でテストすることが多いと思ったので、compose.yaml と Dockerfile を作った。

コンテナイメージとして dind イメージをベースに go を利用できるようにしている。

```
$ docker compose up -d
$ docker compose exec -w /go/src/e2e-framework e2e go test ./sample -v
・
・
=== RUN   TestDeployment
=== RUN   TestDeployment/appsv1/deployment
=== RUN   TestDeployment/appsv1/deployment/deployment_creation
    deployment_test.go:49: deployment found: test-deployment
--- PASS: TestDeployment (15.03s)
    --- PASS: TestDeployment/appsv1/deployment (15.03s)
        --- PASS: TestDeployment/appsv1/deployment/deployment_creation (0.01s)
PASS
ok      github.com/ginoh/e2e-sample/sample      61.371s
```

モジュールのダウンロードのログが初回でるので、`go mod download` を明示的にしておくのでもいいかもしれない。


## kubetest2 との連携

TBD

https://github.com/kubernetes-sigs/e2e-framework/tree/main/third_party/kubetest2


