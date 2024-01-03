### エフェラメラルコンテナを利用したデバッグ

kubernetes 1.23 でエフェメラルコンテナによる debug 利用できる機能が beta になったのでドキュメントを更新

https://kubernetes.io/docs/tasks/debug-application-cluster/debug-running-pod/#ephemeral-container


```
$ minikube -p debug-cluster start --driver hyperkit --kubernetes-version 1.23.0
```

kubectl の version 確認
```
Client Version: version.Info{Major:"1", Minor:"23", GitVersion:"v1.23.0", GitCommit:"ab69524f795c42094a6630298ff53f3c3ebab7f4", GitTreeState:"clean", BuildDate:"2021-12-07T18:08:39Z", GoVersion:"go1.17.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"23", GitVersion:"v1.23.0", GitCommit:"ab69524f795c42094a6630298ff53f3c3ebab7f4", GitTreeState:"clean", BuildDate:"2021-12-07T18:09:57Z", GoVersion:"go1.17.3", Compiler:"gc", Platform:"linux/amd64"}
```

デバッグ対象のシェルも何も入ってないコンテナを立ち上げる
```
$ docker image build -t go-hello .
$ minikube -p debug-cluster image load go-hello:latest
$ kubectl run ephemeral-demo --image=go-hello:latest --restart=Never --image-pull-policy=IfNotPresent
```
exec できない
```
OCI runtime exec failed: exec failed: container_linux.go:380: starting container process caused: exec: "sh": executable file not found in $PATH: unknown
command terminated with exit code 126
```
そこで `kubectl debug`コマンドを実行

```
$ kubectl debug -i -t ephemeral-demo --image=busybox --target=ephemeral-demo
Defaulting debug container name to debugger-bxg8s.
If you don't see a command prompt, try pressing enter.
/ # 
```
`--target` は 別コンテナのプロセス名前空間を指定する

この状態で、`kubectl describe pods`で podの状態を確認してみると、Ephemeral Containers という項目に作成されたコンテナが確認できる。
```
・
・
Containers:
  ephemeral-demo:
    Container ID:   containerd://802cf9fe327fb144844920613b2d07549b4fab56231fa61eede71f2c69a7a150
    Image:          k8s.gcr.io/pause:3.1
    Image ID:       k8s.gcr.io/pause@sha256:f78411e19d84a252e53bff71a4407a5686c46983a2c2eeed83929b888179acea
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Sun, 13 Dec 2020 18:39:12 +0900
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-7tnz5 (ro)
Ephemeral Containers:
  debugger-bxg8s:
    Container ID:   containerd://f49c0a3f8f7eff926bb86e5e08583b8aedc0cce41f12777ccc2c524de9e6b283
    Image:          busybox
    Image ID:       docker.io/library/busybox@sha256:bde48e1751173b709090c2539fdf12d6ba64e88ec7a4301591227ce925f3c678
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Sun, 13 Dec 2020 18:41:06 +0900
    Ready:          False
    Restart Count:  0
    Environment:    <none>
    Mounts:         <none>
・
・    
```

このコンテナは Process 名前空間を共有しているので、デバッグ対象のプロセスを確認できる。

```
/ # ps
PID   USER     TIME  COMMAND
    1 root      0:00 ./hello
   27 root      0:00 sh
   35 root      0:00 ps
/ # 
```
また、 `/proc`経由で ファイルシステムや環境変数を確認できる
```
/ # ls /proc/1/root
dev    etc    hello  proc   sys    var
/ # 
```
このあたりは以下のドキュメントを参照  
https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/

デバッグが終わったら Podは削除
```
$ kubectl delete pods myapp-debug
```
### Pod のコピーを利用したデバッグ

動作中のPodではなく、Podをコピーして、コピーしたPodでデバッグするということも可能

デバッグ対象のコンテナを作成
```
$ kubectl run myapp --image=busybox --restart=Never -- sleep 1d
```
Pod のコピーを作成
```
$ kubectl debug myapp -i -t --image=ubuntu --share-processes --copy-to=myapp-debug

// loginした状態になる
root@myapp-debug:/#

$ kubectl get pods
NAME          READY   STATUS    RESTARTS   AGE
myapp         1/1     Running   0          7m20s
myapp-debug   2/2     Running   0          62s
```

この状態で `myapp-debug`のコンテナを確認してみる。
```
$ kubectl describe pods myapp-debug

・
・
Containers:
  myapp:
    Container ID:  docker://6bf3f3c92dbf36d852db0a44e254d990d4da616530b27030c29f976a67b6b66d
    Image:         busybox
    Image ID:      docker-pullable://busybox@sha256:b5cfd4befc119a590ca1a81d6bb0fa1fb19f1fbebd0397f25fae164abe1e8a6a
    Port:          <none>
    Host Port:     <none>
    Args:
      sleep
      1d
    State:          Running
      Started:      Fri, 10 Dec 2021 23:40:47 +0900
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-v527l (ro)
  debugger-8hpqt:
    Container ID:   docker://4b3061943bfbd17c68d3cf5a88f38ad5a214e2aa92afeea838b477dfb5c8887e
    Image:          ubuntu
    Image ID:       docker-pullable://ubuntu@sha256:626ffe58f6e7566e00254b638eb7e0f3b11d4da9675088f4781a50ae288f3322
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Fri, 10 Dec 2021 23:40:55 +0900
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-v527l (ro)
・
・
```
myapp Pod がコピーされ、コピーされた Podにデバッグ用のコンテナがアタッチされている。

`--share-processes` を利用しているので、myappコンテナのプロセスが見えるようになっている

```
root@myapp-debug:/# ps ax
    PID TTY      STAT   TIME COMMAND
      1 ?        Ss     0:00 /pause
      7 ?        Ss     0:00 sleep 1d
     13 pts/0    Ss     0:00 bash
     37 pts/0    R+     0:00 ps ax
```
/proc/$pid/root/ を利用してコンテナイメージにアクセスも可能
```
root@myapp-debug:/# ls /proc/7/root/    
bin  dev  etc  home  proc  root  sys  tmp  usr  var
```
その他、次のようにすると、Pod をコピーした上で myapp コンテナを sh で実行した状態になる

```
kubectl debug myapp -it --copy-to=myapp-debug --container=myapp -- sh
```
これを利用すると、デバッグ時のみの引数を与えたり、コマンドが間違っていた場合に変更して確認したりができる

また、Podをコピーするときにコンテナイメージを変更することも可能
```
kubectl debug myapp --copy-to=myapp-debug --set-image=*=ubuntu
```
上記は全てのコンテナのイメージを ubuntu に変更している、これを利用するとデバッグしやすいイメージを利用した状態で Pod が起動することになるので、あとはコンテナに入ってデバッグしたりなどができる

```
・
・
Containers:
  myapp:
    Container ID:  docker://b412521497c15508318c087e243028ab4430e50b5f14c3671b67641c1d6def74
    Image:         ubuntu
    Image ID:      docker-pullable://ubuntu@sha256:626ffe58f6e7566e00254b638eb7e0f3b11d4da9675088f4781a50ae288f3322
    Port:          <none>
    Host Port:     <none>
    Args:
      sleep
      1d
    State:          Running
      Started:      Sat, 11 Dec 2021 10:25:30 +0900
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-v527l (ro)
・
・
```

デバッグが終わったら Podは削除
```
$ kubectl delete pods myapp-debug
```

Node 上の shell を通してのデバッグ

(確認中)
```
$ kubectl get nodesNAME                STATUS   ROLES                  AGE     VERSION
debug-cluster       Ready    control-plane,master   2d13h   v1.23.0
debug-cluster-m02   Ready    <none>                 2d13h   v1.23.0


$ kubectl debug node/debug-cluster-m02 -i -t --image=ubuntu

// node の filesystem は /host にマウントされている
root@debug-cluster-m02:/# ls /host
bin   etc   lib      linuxrc  opt   run   sys  var
data  home  lib64    media    proc  sbin  tmp
dev   init  libexec  mnt      root  srv   usr

// デバッグコンテナの削除
$ kubectl delete pods node-debugger-debug-cluster-m02-kztvf
```