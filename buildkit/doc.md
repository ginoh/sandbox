### buildkit

環境準備

```
$ minikube start -p buildkit --driver hyperkit
$ minikube addons enable registry -p buildkit
$ wget https://raw.githubusercontent.com/moby/buildkit/master/examples/kubernetes/pod.rootless.yaml
$ kubectl apply -f pod.rootless.yaml

// install buildctl
brew install buildkit
```

SBOM に関するドキュメント
https://github.com/moby/buildkit/blob/master/docs/attestations/sbom.md

SBOM 生成は以下のプロトコルにしたがう (SBOM Scanning Protocol)
https://github.com/moby/buildkit/blob/master/docs/attestations/sbom-protocol.md

デフォルトスキャナーとして以下のイメージが実装されている
* docker/buildkit-syft-scanner

```
$ cd sample-image
$ buildctl \
  --addr kube-pod://buildkitd \
  build --frontend dockerfile.v0 \
  --output type=image,name=registry.kube-system.svc.cluster.local/sample-image,push=true,registry.insecure=true \
  --local context=. \
  --local dockerfile=. \
  --opt attest:sbom=
```
`attest:sbom=` はデフォルトの指定

SBOM Scanning Protocol 必須のパラメータなどは buildkit で設定されているように見えた
https://github.com/moby/buildkit/blob/master/frontend/attestations/sbom/sbom.go#L23
https://github.com/moby/buildkit/blob/master/frontend/attestations/sbom/sbom.go#L61

```
$ curl http://$(minikube -p buildkit ip):5000/v2/_catalog
{"repositories":["sample-image"]}
```

`docker buildx imagetools inspect` でレジストリにアクセスする際に、`http` でアクセスできるように準備する

`docker buildx` でインスタンスを作った上で、buildkitd の config を読みこませる
https://docs.docker.com/build/buildkit/configure/
https://docs.docker.com/build/buildkit/toml-configuration/

```
$ minikube ip -p buildkit                                          
192.168.64.78

$ cat buildkitd.default.toml
[registry."192.168.64.78:5000"]
  http = true


$ docker buildx create --name buildkit-sbom --bootstrap
WARNING: Using default BuildKit config in /Users/hsugino/.docker/buildx/buildkitd.default.toml
[+] Building 8.0s (1/1)FINISHED
・
・
・
buildkit-sbom

$ docker buildx inspect buildkit-sbom
Name:   buildkit-sbom
Driver: docker-container

Nodes:
Name:      buildkit-sbom0
Endpoint:  desktop-linux
Status:    running
Buildkit:  v0.11.2
Platforms: linux/amd64, linux/amd64/v2, linux/amd64/v3, linux/amd64/v4, linux/arm64, linux/riscv64, linux/ppc64le, linux/s390x, linux/386, linux/mips64le, linux/mips64, linux/arm/v7, linux/arm/v6


$ docker buildx imagetools --builder buildkit-sbom inspect $(minikube -p buildkit ip):5000/sample-image
Name:      192.168.64.78:5000/sample-image:latest
MediaType: application/vnd.oci.image.index.v1+json
Digest:    sha256:362a29542686e022d487d74efcc596b3bd2d027ad57842799315cc8225e651c4

Manifests:
  Name:      192.168.64.78:5000/sample-image:latest@sha256:6afe0e0cbdc543bb851d98fc624e0789ce070f319afbdaf7833f15d15ee62cd4
  MediaType: application/vnd.oci.image.manifest.v1+json
  Platform:  linux/amd64

  Name:      192.168.64.78:5000/sample-image:latest@sha256:3b21dde0914a95ee4555d41c2f46d3c84e0068530a9dcce36fa4aae93695345a
  MediaType: application/vnd.oci.image.manifest.v1+json
  Platform:  unknown/unknown
    vnd.docker.reference.digest: sha256:6afe0e0cbdc543bb851d98fc624e0789ce070f319afbdaf7833f15d15ee62cd4
    vnd.docker.reference.type:   attestation-manifest
```

色々設定したけど、`docker buildx imagetools inspect` でやっているのは結局これと同じだった (これの方が楽)
```
$ curl -H "Accept: application/vnd.oci.image.index.v1+json" http://$(minikube -p buildkit ip):5000/v2/sample-image/manifests/latest
```

```
$ curl -H "Accept: application/vnd.oci.image.index.v1+json" http://$(minikube -p buildkit ip):5000/v2/sample-image/manifests/latest
{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "digest": "sha256:37e3f0b18c5aaf222d53a8e992dc1639f301fff83c233112ab48a613abb64df4",
    "size": 167
  },
  "layers": [
    {
      "mediaType": "application/vnd.in-toto+json",
      "digest": "sha256:1c6dc80de20a87622e70b3cbe24492f5f46ce446ec7ac2cf749ac022d00e31b5",
      "size": 54306,
      "annotations": {
        "in-toto.io/predicate-type": "https://spdx.dev/Document"
      }
    }
  ]
}
```

pull できる？

buildkitd の設定で `http = true` がなにしているかは　buildkit のこの辺りのコードをみるとよい
https://github.com/docker/buildx/blob/4c938c77bab00dbe983597b7d29f2e2eb19e8000/util/resolver/resolver.go#L31


#### その他、気になる更新 (別途見る予定)
* BuildKit now supports reproducible builds by setting SOURCE_DATE_EPOCH build argument or source-date-epoch exporter attribute. This deterministic date will be used in image metadata instead of the current time.

* New Build History API allows listening to events about builds starting and completing, and streaming progress of active builds. New commands buildctl debug monitor, buildctl debug logs and buildctl debug get have been added to use this API. Build records also keep OpenTelemetry traces, provenance attestations, and image manifests if they were created by the build.

* Setting multiple cache exporters for a single build is now supported
* Remote cache import/export to client-side local files now supports tag parameter for scoping cache
* RegistryToken auth from Docker config is now allowed as authentication input
* buildctl now loads Github runtime environment when using GHA remote cache
