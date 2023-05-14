# 参考

https://cilium.io/
https://www.publickey1.jp/blog/23/ebpfebpf_cloud_native_days_tokyo_2022.html (CNDT 2022)

# Cilium とは

Cilium は Networking、 Security、Observability を kubernetes のような Cloud Native 環境に導入するオープンソースプロジェクト。
eBPF と呼ばれる Linux kernel テクノロジを基盤として、セキュリティ、可視性、ネットワーク制御ロジックを動的に挿入できる。

## Networking

Service Load Blancing
Scalable Kubernetes CNI
Multi-cluster Connectivity

## Observability

TBD

## Security

TBD

# Architecture

https://cilium.io/get-started/ の Architecture 参照

ノードやサーバにエージェントを入れるらしい。

# GETTING STARTED

https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/

```
$ minikube -p cilium-sample start --driver qemu --network socket_vmnet --network-plugin=cni --cni=false
```

minikube には `--cni` で cilium の導入が簡単にできるが、cilium の version が低いので上記のようにして自分で cilium いれるらしい

CLI インストール
https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/#install-the-cilium-cli

Cilium インストール

```
$ kubectl config current-context
cilium-sample

$ cilium install
ℹ️  Using Cilium version 1.13.2
🔮 Auto-detected cluster name: cilium-sample
🔮 Auto-detected datapath mode: tunnel
🔮 Auto-detected kube-proxy has been installed
ℹ️  helm template --namespace kube-system cilium cilium/cilium --version 1.13.2 --set cluster.id=0,cluster.name=cilium-sample,encryption.nodeEncryption=false,kubeProxyReplacement=disabled,operator.replicas=1,serviceAccounts.cilium.name=cilium,serviceAccounts.operator.name=cilium-operator,tunnel=vxlan
ℹ️  Storing helm values file in kube-system/cilium-cli-helm-values Secret
🔑 Created CA in secret cilium-ca
🔑 Generating certificates for Hubble...
🚀 Creating Service accounts...
🚀 Creating Cluster roles...
🚀 Creating ConfigMap for Cilium version 1.13.2...
🚀 Creating Agent DaemonSet...
🚀 Creating Operator Deployment...
⌛ Waiting for Cilium to be installed and ready...
    /¯¯\
 /¯¯\__/¯¯\    Cilium:          1 errors, 1 warnings
 \__/¯¯\__/    Operator:        OK
 /¯¯\__/¯¯\    Hubble Relay:    disabled
 \__/¯¯\__/    ClusterMesh:     disabled
    \__/

Deployment        cilium-operator    Desired: 1, Ready: 1/1, Available: 1/1
DaemonSet         cilium             Desired: 1, Unavailable: 1/1
Containers:       cilium             Pending: 1
                  cilium-operator    Running: 1
Cluster Pods:     0/1 managed by Cilium
Image versions    cilium             quay.io/cilium/cilium:v1.13.2@sha256:85708b11d45647c35b9288e0de0706d24a5ce8a378166cadc700f756cc1a38d6: 1
                  cilium-operator    quay.io/cilium/operator-generic:v1.13.2@sha256:a1982c0a22297aaac3563e428c330e17668305a41865a842dec53d241c5490ab: 1
Errors:           cilium             cilium          1 pods of DaemonSet cilium are not ready
Warnings:         cilium             cilium-2nxp7    pod is pending
↩️ Rolling back installation...

Error: Unable to install Cilium: timeout while waiting for status to become successful: context deadline exceeded
```

なんかエラーになった。もう一回試したところ、daemonset の Pod を立ち上げるときに以下の event が確認できた。

```
  Warning  FailedMount  68s (x10 over 5m18s)  kubelet            MountVolume.SetUp failed for volume "bpf-maps" : mkdir /sys/fs/bpf: operation not permitted
```

ドキュメントにクラスタ作る際の注意点として以下があり、minikube の version は v1.30.1 だったので無視していた。

```
MacOS M1 users using a Minikube version < v1.28.0 with --cni=false will also need to run minikube ssh -- sudo mount bpffs -t bpf /sys/fs/bpf in order to mount the BPF filesystem bpffs to /sys/fs/bpf.
```

実行したがダメだった。

```
minikube -p cilium-sample ssh -- sudo mount bpffs -t bpf /sys/fs/bpf
mount: /sys/fs/bpf: mount point does not exist.
ssh: Process exited with status 32
```

これか？
https://github.com/kubernetes/minikube/issues/14674

ちなみに、`kubectl -n kube-system get pods` したときに、coredns と storage-provisioner が pending 状態で起動してなかった。

しかたないので、qemu driver を使うのを一旦やめた

```
$ minikube -p cilium-sample-docker start --driver docker --network-plugin=cni --cni=false
$ cilium install
$ cilium status
    /¯¯\
 /¯¯\__/¯¯\    Cilium:          OK
 \__/¯¯\__/    Operator:        OK
 /¯¯\__/¯¯\    Hubble Relay:    disabled
 \__/¯¯\__/    ClusterMesh:     disabled
    \__/

DaemonSet         cilium             Desired: 1, Ready: 1/1, Available: 1/1
Deployment        cilium-operator    Desired: 1, Ready: 1/1, Available: 1/1
Containers:       cilium-operator    Running: 1
                  cilium             Running: 1
Cluster Pods:     1/1 managed by Cilium
Image versions    cilium             quay.io/cilium/cilium:v1.13.2@sha256:85708b11d45647c35b9288e0de0706d24a5ce8a378166cadc700f756cc1a38d6: 1
                  cilium-operator    quay.io/cilium/operator-generic:v1.13.2@sha256:a1982c0a22297aaac3563e428c330e17668305a41865a842dec53d241c5490ab: 1


$ cilium connectivity test
・
・
・
........

✅ All 41 tests (180 actions) successful, 8 tests skipped, 1 scenarios skipped.
```
