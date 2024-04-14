# DinD コンテンを権限を弱めて利用できるかを調べる

https://hub.docker.com/_/docker

dind を priviledge なしで起動してみる

=> 次のエラーログで起動しない

```
 docker container run --name dind -d docker:dind

Certificate request self-signature ok
subject=CN = docker:dind server
/certs/server/cert.pem: OK
Certificate request self-signature ok
subject=CN = docker:dind client
/certs/client/cert.pem: OK
ip: can't find device 'ip_tables'
modprobe: can't change directory to '/lib/modules': No such file or directory
mount: permission denied (are you root?)
Could not mount /sys/kernel/security.
AppArmor detection and --privileged mode might break.
mount: permission denied (are you root?)
```

rootless の時も privileged 指定していない時は同じ。
dind script のエラー
https://github.com/docker-library/docker/blob/679ff1a932a1bdf4341fb0d5cb06f088c77d44bf/24/dind/dockerd-entrypoint.sh#L189
https://raw.githubusercontent.com/docker/docker/1f32e3c95d72a29b3eaacba156ed675dba976cb5/hack/dind

フラグをつけて起動したときは以下のログがでる

```
Certificate request self-signature ok
subject=CN = docker:dind server
/certs/server/cert.pem: OK
Certificate request self-signature ok
subject=CN = docker:dind client
/certs/client/cert.pem: OK
Device "ip_tables" does not exist.
modprobe: can't change directory to '/lib/modules': No such file or directory
```

DinD について須田さんの書籍にかいてあった内容

- dind は privileged フラグが必要
- rootless モードを使うと 非 root 権限で実行可能だが、privileged 自体は引き続き必要
- 実験的プロジェクト Docker-in-UML-in-Docker が privileged が不要だがオーバーヘッドがかなりでかい、あと最近の更新がなさそう？
- containerd 1.3 以降を使う場合は、privileged_without_host_devices モードがある

参考になりそうな blog
https://www.docker.com/blog/docker-can-now-run-within-docker/
https://qiita.com/muddydixon/items/d2982ab0846002bf3ea8
https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities

Tekton の dind 利用について

DinD コンテナを privileged 指定せずに実行できるか？
=> 不可能と思って良さそう

privilege でも安全な方法をとれるか？
=> 以下を組み合わせるとよいかもしれない

- rootless dokcer で権限をしぼる
- privileged 権限は起動時に必要なだけ？
- tls
- 特定のコンテナからのみアクセス可能とする
- もしくは、tls を閉じる

参考メモ
mac の Docker VM に入る方法

Mac VM login
https://qiita.com/notakaos/items/b08ba7166bb5b56576a1
https://zenn.dev/mythrnr/scraps/71a11550d90024

cgroup
https://qiita.com/ymktmk/items/d8f3d000325608e714c2
https://www.itbook.info/network/docker06.html
https://gihyo.jp/admin/serial/01/linux_containers/0004

VM の中

```
$ docker run -it --rm --privileged --pid=host alpine:edge nsenter -t 1 -m -u -n -i sh
/ # ls -l /proc/1/cgroup
-r--r--r--    1 root     root             0 May 28 05:46 /proc/1/cgroup
/ # cat /proc/1/cgroup
0::/../..
```

rootless-docker
=> userns によってコンテナ内の root は ホストの root 権限をもたない
=> docker daemon も root 権限をもたない

best practice
https://docs.docker.com/engine/security/rootless/#rootless-docker-in-docker

best practice の内容が自動で設定されている
https://github.com/docker-library/docker/blob/master/24/dind/dockerd-entrypoint.sh#L117

https://tekton.dev/vault/pipelines-main/tasks/#specifying-sidecars
