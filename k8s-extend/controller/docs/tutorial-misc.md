つまづいた部分のメモ

kb init する際に go 1.17 じゃないと controller-gen のバイナリが生成されない
local でだけ .go-versionがあっても、最初のコマンド実行時はいいが 内部で、go get する際にはみていなさそう
なので、globalを1.17にしておく必要がありそう


Makefile 実行時に バイナリを取得する 処理がうまく動かなかった
=> go-get-tool が tmpdirで作業している。
tmpdirなので、goenv で生成された .go-versionがなく、global設定を参照する
go 1.18.0が globalだと go get ではコマンドはインストールされない

workaround
```
# workaround
# use go1.18 with global + use 1.17 and below with local
GOENV_GOVERSION := $(shell goenv version-name)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
echo "goenv local ${GOENV_GOVERSION}" ;\
goenv local ${GOENV_GOVERSION} ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get -v $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
```

config/manager/manager.yaml  に imagePullPolicy を設定
```
minikube -p sandbox-controller --driver hyperkit start -n 3
// https://cert-manager.io/docs/installation/
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
make docker-build
minikube -p sandbox-controller image load controller:latest
```

brew install だと kubebuilder が依存関係で go もいれてしまう。go 1.18.1 が使われないようにしたい
kubebuilder は brew でいれるのやめた
