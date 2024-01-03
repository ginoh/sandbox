### 概要

Local環境に Kubernetes クラスタを作成するツールである minikube を利用する際のメモ

### Requirement
* 2 CPUs or more
* 2GB of free memory
* 20GB of free disk space
* Internet connection
* Container or virtual machine manager, such as:
 - Docker
 - Hyperkit
 - Hyper-V
 - KVM
 - Parallels
 - Podman
 - VirtualBox
 - VMware Fusion/Workstation
### 参考

minikube の install
https://kubernetes.io/ja/docs/tasks/tools/install-minikube/
https://kubernetes.io/ja/docs/setup/learning-environment/minikube/

Hello Minikube
https://kubernetes.io/ja/docs/tutorials/hello-minikube/




### セットアップ

今回は Mac での利用を想定
#### ハイパーバイザのインストール

Intel Mac を使用していて、VMとして Hyperkit を利用したい場合、Docker for Mac をインストールしていると　Hyperkitは存在しているので別途インストールなどする必要はない

```
$ which hyperkit
/usr/local/bin/hyperkit

$ ls -l /usr/local/bin/hyperkit 
/usr/local/bin/hyperkit@ -> /Applications/Docker.app/Contents/Resources/bin/com.docker.hyperkit
```

Apple Silicon Mac を使用していて、VM として qemu を利用したい場合、brew でインストールする
```
$ brew install qemu
```
qemu を利用したときの Networking として `socket_vmnet` のインストールをしておいたほうがよさそう
https://minikube.sigs.k8s.io/docs/drivers/qemu/#networking

`service` や `tunnel`　を使うためには `socket_vmnet` が必要だが、それ以外にも `minikube start` で複数のクラスタを作る際に `--network` フラグで `builtin` の指定や何も指定しないと IP に同じものが割り振られてしまってうまく動かなかった


#### minikube のインストール

```
$ brew install minikube
```

確認

Intel Mac の場合
```
$ minikube start --driver hyperkit
$ minikube status
minikube
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured

$ minikube delete
```
マルチノードクラスタを作成したい時は、`-n` フラグをつける

Apple Silicon Mac の場合

```
$ minikube start --driver qemu --network socket_vmnet
$ minikube status
minikube
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured
```

### minikube Quick Start

クイックスタートをやってみる

```
$ minikube start 
$ kubectl create deployment hello-minikube --image=k8s.gcr.io/echoserver:1.10
$ kubectl expose deployment hello-minikube --type=NodePort --port=8080
$ kubectl get pods,svc
NAME                                  READY   STATUS    RESTARTS   AGE
pod/hello-minikube-5d9b964bfb-q5gq6   1/1     Running   0          9m15s

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/hello-minikube   NodePort    10.104.207.93   <none>        8080:30168/TCP   117s
service/kubernetes       ClusterIP   10.96.0.1       <none>        443/TCP          10m
```

以下のコマンドでアクセスURLが取得できる
```
minikube service hello-minikube --url
```

```
$ kubectl delete services hello-minikube
$ kubectl delete deployment hello-minikube
$ minikube stop
$ minikube delete
```

### minikube クラスタ内で利用するコンテナイメージの追加方法

参考：
* https://minikube.sigs.k8s.io/docs/handbook/pushing/

個人的に使いそうなものをピックアップ

1. クラスタ内の Docker daemon に対して push する

ローカルの docker client のアクセス先をクラスタ内の Docker daemonにすることで、ローカルでビルドしたコンテナイメージがそのままクラスタ内で利用できる

アクセス先の切り替え (Mac)
```
eval $(minikube docker-env)
```
あとは、普通に docker image build する

2. Registryアドオンを利用してクラスタ内部に Registry を構築しコンテナイメージを push する

参考: 
* https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon
* https://github.com/kubernetes/minikube/blob/v1.30.1/pkg/minikube/cluster/ip.go#L37 (上記ページのリンク先が古いため)

minikube には Registry用の addonがあるのでそれを利用する
```
$ minikube addons list

|-----------------------------|-----------------------|
|         ADDON NAME          |      MAINTAINER       |
|-----------------------------|-----------------------|
| ambassador                  | unknown (third-party) |
・
・
| registry                    | google                |
| registry-aliases            | unknown (third-party) |
| registry-creds              | unknown (third-party) |
```

利用方法としては以下になる
* kubernetes クラスタを insecure registry flag を利用して起動する
* registry addon を有効にする
* ローカルのコンテナイメージを扱う client で insecureの設定をする

```
// insecure-registryに何を指定するかは環境によって異なる
$ minikube -p registry-sample start  -n 2 --driver hyperkit --insecure-registry "192.168.64.0/24"

// kube-system に registry 関連の pod とか立ち上がる
$ minikube -p registry-sample addons enable registry

// https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L435
// を参考に Docker の insecure registryを設定する

// イメージビルドと push
$ docker image build -t $(minikube ip -p registry-sample):5000/test-image .
$ docker push $(minikube ip -p registry-sample):5000/test-image

// 適当な deployment作って試す
```


3. クラスタ内のコンテナランタイムへのコンテナイメージのロード

参考：
* https://minikube.sigs.k8s.io/docs/commands/image/#minikube-image-load

```
// e.g. local でイメージをビルド
$ docker image build -t test-image .

$ minikube -p registry-sample image load test-image

// 適当な deployment作って試す
```
imagePullPolicyを IfNotPresent にしておく

### マルチノードで PersistentVolume を使う方法

デフォルトではマルチノードクラスターを構築した際、PersistentVolume は正常に動作しないため `CSI Hostpath Driver` addon を利用する

参考:
* https://minikube.sigs.k8s.io/docs/tutorials/multi_node/#caveat
* https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/

```
$ minikube -p csi-hostpath-driver-demo start --driver qemu --network socket_vmnet -n 2

$ minikube -p csi-hostpath-driver-demo addons enable volumesnapshots
$ minikube -p csi-hostpath-driver-demo addons enable csi-hostpath-driver

// Optional. For dynamic volume claim.
$ minikube -p csi-hostpath-driver-demo addons disable storage-provisioner
$ minikube -p csi-hostpath-driver-demo addons disable default-storageclass
$ kubectl patch storageclass csi-hostpath-sc -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```
以下の Tutorial をやるとよい
https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/#tutorial