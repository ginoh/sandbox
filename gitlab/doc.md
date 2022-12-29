## Gitlab

ローカルの k8s クラスタに GitLab を構築する時にしたことのメモ


### k8s クラスタ構築

```
$ minikube -p gitlab start --driver hyperkit
$ kubectl create namespace gitlab-test
```

### gitlab 構築

k8s 上での構築には [Helm Chart](https://docs.gitlab.com/charts/installation/deployment.html#deploy-using-helm) を利用する方法があるが、この場合コンポーネントごとのコンテナイメージを利用してマイクロサービス化した構成になるようだ。対して [Docker](https://docs.gitlab.com/ee/install/docker.html) は オールインワンコンテナイメージである gitlab-ce を利用する方法。

今回は一つのコンテナで GitLab を構築する方法を取る。

#### パスワードの生成

```
$ kubectl -n gitlab-test create secret generic gitlab-initial-root-password --from-literal=password=$(head -c 512 /dev/urandom | LC_CTYPE=C tr -cd 'a-zA-Z0-9' | head -c 32)
```

#### gitlab リソース適用

```
$ kubectl -n gitlab-test apply -f ./gitlab.yaml 
```

#### その他

`GITLAB_OMNIBUS_CONFIG` などで必要な設定を k8s リソースの environment で設定する