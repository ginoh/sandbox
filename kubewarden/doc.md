kubewarden は kubernetes の Dynamic Admission Control
の仕組みを利用して Policy as Code を実現するソフトウェア

同様のソフトウェアとしては OPA (Open Policy Agent) や kyverno が存在する

その特徴として、仕組み的に kubernetes 環境の実行が前提ではあるが、policy のロジック実行を WebAssembly で実現するため、WebAssembly を生成できるプログラミング言語で記述できる

https://github.com/kubewarden

policy は OCI Artifact として OCI レジストリや HTTP Server で配布可能

### QuickStart

### install

cert-manager

```
kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
```

kubewarden

```
helm repo add kubewarden https://charts.kubewarden.io

helm install --wait -n kubewarden --create-namespace kubewarden-crds kubewarden/kubewarden-crds

helm install --wait -n kubewarden kubewarden-controller kubewarden/kubewarden-controller
```

```
helm list -A
NAME                    NAMESPACE       REVISION        UPDATED                                     STATUS          CHART                           APP VERSION
kubewarden-controller   kubewarden      1               2022-02-23 22:13:33.953041 +0900 JST        deployed        kubewarden-controller-0.3.6     v0.4.5
kubewarden-crds         kubewarden      1               2022-02-23 22:13:13.823851 +0900 JST        deployed        kubewarden-crds-0.1.1
```

### Example

kubewarden をインストールすると、default の policy-server が立ち上がっている

```
$ kubectl -n kubewarden get pods
NAME                                     READY   STATUS    RESTARTS      AGE
kubewarden-controller-74bbf45677-xc6kq   1/1     Running   2 (34h ago)   3d
policy-server-default-86b8758987-p8vr4   0/1     Running   0             47h
```

この policy-server は Cluster スコープのカスタムリソースとして登録されている

```
$ kubectl get policyservers
NAME      AGE
default   3d1h
```

policy として ClusterAdmissionPolicy を登録する

```
$ kubectl apply -f policy/privileged-pods.yaml

$ cat policy/privileged-pods.yaml
apiVersion: policies.kubewarden.io/v1alpha2
kind: ClusterAdmissionPolicy
metadata:
  name: privileged-pods
spec:
  module: registry://ghcr.io/kubewarden/policies/pod-privileged:v0.1.9
  rules:
    - apiGroups: [""]
      apiVersions: ["v1"]
      resources: ["pods"]
      operations:
        - CREATE
        - UPDATE
  mutating: false
```

ClusterAdmissionPolicy では policy-server と実行する policy module を指定する

上記の例では default の policy-server を利用し、指定された policy をダウンロードし実行するようになる (controller が処理する)

```
$ kubectl get clusteradmissionpolicies
```

登録直後は status が pending になるが、webhook が登録されると、status が active になる

```
$ kubectl get validatingwebhookconfigurations
```

今回の例では priviledge の権限を与えた pod の実行を拒否するものなので以下を適用して実行できないことを確認

```
$ kubectl apply -f privileged-pod.yaml
```

policy-server はカスタムリソースを登録することで 独自に deployment を登録することが可能

この仕組みにより以下のような利点が得られる

- policy が多い namespace などの処理を分離することが可能
- 重要な policy を扱う server を server pool 内で処理可能になるため障害に強くなる

```
$ kubectl apply -f server/reserved-instance.yaml
```

### Architecture

https://docs.kubewarden.io/architecture.html

- kubewarden は以下の コンポーネントで構成される
- kubernetes カスタムリソース
- kabernetes カスタムコントローラ
- policy (WebAssembly モジュール)
- policy-server
