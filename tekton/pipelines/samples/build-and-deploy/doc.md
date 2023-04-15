## 注意点/わかったこと
### minikube

* minikube は デフォルトでは control-plane 以外で Persistent Volume を使おうとすると permission が 777 ではなく 755 になるよう
https://github.com/kubernetes/minikube/issues/12360
  * Tekotn workspace(Persistent Volume) に対して、データを保存するときに、non-root なユーザで実行しようとすると
Permission Error になった

* in-cluster の Registry から image の pull をする際、イメージは VM からの pull になるため、(VMのIP):5000 でアクセスする必要がある
  * クラスタ内での push => registry.kube-system.svc.cluster.local/imageName
  * image の pull => (VMのIP):5000/imageName


### Buildkit

rootlessモード

https://github.com/moby/buildkit/blob/master/docs/rootless.md

* 環境変数でフラグ(`--oci-worker-no-process-sandbox`)を指定すると、privileged の設定は必要なくなる
* seccomp, apparmor の無効化は必要 (minikube の場合は気にしなくてもいい？)
* GKE の COS で rootless モードを動かす場合は以下を参考にする
  * https://github.com/moby/buildkit/pull/3097

insecure Registry

output フラグで `registry.insecure=true` を指定すると insecure なレジストリにイメージを push できるが、
export-cache フラグで `type=registry` にする場合、https のアクセスになってしまうよう
* https://github.com/moby/buildkit/issues/2054
* https://github.com/moby/buildkit/issues/2044


minikube 環境だと、apparmor/seccomp の無効設定しなくても動いた

```
buildctl-daemonless.sh --debug build \
--progress plain \
--frontend dockerfile.v0 \
--opt filename=Dockerfile \
--local context=. \
--local dockerfile=. \
--output type=image,name=registry.kube-system.svc.cluster.local/sample-hello-world,push=true,registry.insecure=true \
--export-cache type=inline \
--import-cache type=registry,ref=registry.kube-system.svc.cluster.local/sample-hello-world:buildcache
```
###  その他

最終的に関係なかったが後で思い出すためのメモ

k8s v1.24 からは サービスアカウントの Token はバウンドサービスアカウントトークンが利用され、Secret の Token は作成されなくなった

https://qiita.com/uesyn/items/90ca3789ce88cb9ea7a4#%E3%82%A2%E3%83%83%E3%83%97%E3%82%B0%E3%83%AC%E3%83%BC%E3%83%89%E6%99%82%E3%81%AE%E6%B3%A8%E6%84%8F%E4%BA%8B%E9%A0%85

kubeconfig でサービスアカウントトークンが使いたかったが、検証程度なら期限ありのToken でも問題ないので作った
```
kubectl -n <namespace> create token <service account name> --duration <duration time (sec)> > sa-token
// e.g. kubectl -n sample-build-and-deploy create token kubectl-apply --duration 31536000s > sa-token
// duration は API serverの設定により、最小・最大がありそう

kubectl --kubeconfig sample-kubeconfig config set-cluster tekton-sandbox --insecure-skip-tls-verify=true --server=https://192.168.64.61:8443
$ kubectl --kubeconfig sample-kubeconfig config set-credentials kubectl-apply --token $(cat sa-token)
$ kubectl --kubeconfig sample-kubeconfig config set-context sample-build-and-deploy --cluster tekton-sandbox --user kubectl-apply
```



