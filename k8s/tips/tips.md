個人的な細かい Memo など

### dryrunを利用してresourceを作成する雛形にする
kubectlの `--dry-run`オプションを利用することで resourceを作成する際の雛形として利用する

参考: [generator](https://kubernetes.io/docs/reference/kubectl/conventions/#generators)

(メモ)  
以前は、`kubectl run`を利用した雛形作成の情報がいたるところにあるが、`run-pod/v1`以外の`generator`が非推奨になっており pod作成以外は`kubectl create`のdryrunをしたほうがよいと思われる。

Podの作成
```
$ kubectl run nginx --generator=run-pod/v1 --image nginx:1.13 --dry-run -o yaml
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  labels:
    run: nginx
  name: nginx
spec:
  containers:
  - image: nginx:1.13
    name: nginx
    resources: {}
  dnsPolicy: ClusterFirst
  restartPolicy: Always
status: {}
```
`--dry-run`の結果として出力された内容を編集してresourceのためのyamlを作成して利用する

deploymentの作成の場合
```
$ kubectl create  deployment nginx --image=nginx:1.13 --dry-run -o yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: nginx
  name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx:1.13
        name: nginx
        resources: {}
status: {}
```
この方法で yamlを生成した際、配列を値としてもつハッシュは以下のように表現されていた
```
containers:
- image: nginx:1.13
```
上記の例でいうところのimageのインデントの位置が気になる人はインデントをいれたり何かしらのFormatterツールを利用して整形するとよい
例えば、Prettier でフォーマットすると以下のようになる
```
containers:
  - image: nginx:1.13
```

### kubectl port-forward を利用してクラスタ外からクラスタ内のPodにアクセスする
nginx の pod を作成
```
$ kubectl run nginx --image nginx --port 80
```

port-forward を利用して local から port-forward
```
$ kubectl port-forward pod/nginx 8080:80
kubectl port-forward nginx  8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80

// 別の terminal など
$ curl -s localhost:8080 | head -n 5
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
```
ちなみにPod起動時に `--port 80` がなくてもつながる。  
`--port` はメタデータ保存してるようなもの？ Dockerの expose的な？ (今後調べる)

port-forard は serviceを指定可能

nginx の deployment を 作成
```
$ kubectl create deployment nginx --image nginx --port 80 --replicas=3 --dry-run=client -o yaml | kubectl apply -f -
```

expose を利用して service をつくる
```
$ kubectl expose deployments nginx
service/nginx exposed

// --port で exposeしていれば、expose コマンド実行時に port指定しなくてもいい
$ kubectl get services nginx
NAME    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
nginx   ClusterIP   10.96.190.36   <none>        80/TCP    41s
```

port-forward
```
$ kubectl port-forward services/nginx 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80

// 別のterminal など
$ curl -s localhost:8080 | head -n 5
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
```
注意. service を利用してport-forard はしているが、動作としては特定の一つのPodだけと通信している  
(-v オプションで ログなどを見るとわかる。)

### kubectl proxy を利用してクラスタ内のServiceにアクセスする

TBD