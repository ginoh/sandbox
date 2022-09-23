### install

```
kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
kubectl apply --filename https://storage.googleapis.com/tekton-releases/triggers/latest/interceptors.yaml
```

### Getting Started

[Tutorial: Getting started with Tekton Triggers](https://github.com/tektoncd/triggers/tree/main/docs/getting-started) を参考にしようと思ったが古そうなので独自に変更

気になった点
* `/tekton/home` が存在する。`/tekton/home` ディレクトリは内部仕様であり使うべきではない
* Deprecated な Pipeline Resource が使われている

変更した点
* Pipeline
* 今回は ingress は使わない


memo: 
Pipeline Resource の cloudevent は task で置き換える
https://github.com/tektoncd/catalog/tree/main/task/cloudevent/0.1/samples

また、現在は Taskごとの sinkの設定は対応せず Tekton 全体の config で設定するのみ
https://github.com/tektoncd/pipeline/issues/2082

https://github.com/tektoncd/pipeline/blob/main/docs/auth.md


```
kubectl create ns getting-started
kubectl -n getting-started apply -f rbac/
kubectl -n getting-started apply -f pipelines/
kubectl -n getting-started apply -f triggers.yaml
kubectl -n  getting-started apply -f ingress-webhook/task-create-webhook.yaml
```

[ngrok](https://ngrok.com/) と  Service(NodePort), kubectl port-forward、ingress-dns (minikube addon)
などを組み合わせて外部からの疎通を確保する

今回は ngrok + service (type=NodePort)
```
kc -n getting-started expose service el-getting-started-listener --type NodePort --name el-getting-started-listener-nodeport

// 8080 のアクセスの方を外部から疎通できるように使う
minikube -p tekton-sandbox service list
minikube -p tekton-sandbox service -n getting-started el-getting-started-listener-nodeport --url
```

webhook を作るが、別に手動でやってもいい
```
// github リポジトリで webhook を作成するための Token を作る
// secret を replace した後に適用
kubectl -n  getting-started apply -f webhook/secret.yaml

// service の 8080に疎通でき URL を指定
ngrok http xxxxxx

// ExternalDomain を ngrok の ドメインにする
kubectl -n getting-started apply -f webhook/run-task-create-webhook.yaml
```

GCP SA の json をとってきて Secret つくっておく (image の push 用)
```
kubectl create secret generic gcp-service-account --from-file=service-account=xxxx.json -n getting-started --dry-run=client -o yaml
```

GCP SA の json をとってきて imagePullSecret を設定する (deploy 後の image pull 用)
```
// https://cloud.google.com/artifact-registry/docs/access-control?hl=ja#pullsecrets
kubectl create secret docker-registry artifact-registry \
--docker-server=https://LOCATION-docker.pkg.dev \
--docker-email=SERVICE-ACCOUNT-EMAIL \
--docker-username=_json_key \
--docker-password="$(cat KEY-FILE)"

kubectl edit serviceaccount default --namespace xxxxx

imagePullSecrets:
- name: artifact-registry
```

GitHub リポジトリで適当なブランチにコミットする 
```
git commit -a -m "build commit" --allow-empty && git push origin tekton-triggers
```

memo:

docker-login Task で コンテナが root user だと config.json の permission が root:root になる

そのため rootless モードの buidkit で config.json を読み込むときに permission エラーになる