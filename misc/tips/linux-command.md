### strace/ltrace/ptrace

Linuxでアプリケーションをうまく動作しないときのデバッグ方法として,strace/ltraceなどのコマンドを使う時がある。

どういったものか忘れてしまうことが多いのでメモをしておく。


#### strace

straceは特定プロセスのシステムコールの呼び出し履歴を調べることができる。

システムコールレベルでの処理を調べることで不具合などの問題調査の役に立つ。

利用準備 

```
// straceのインストール
$ yum install strace

// docker コンテナ内で利用する場合は --cap-add=SYS_PTRACE を利用する
// e.g.
$ docker container run --cap-add=SYS_PTRACE --name centos --rm -i -t centos zsh
$ yum install strace

```

利用例
```
strace ls /bin
execve("/usr/bin/ls", ["ls", "/bin"], 0x7ffcb93b2468 /* 9 vars */) = 0
brk(NULL)                               = 0x561d19ff9000
arch_prctl(0x3001 /* ARCH_??? */, 0x7ffe75040e80) = -1 EINVAL (Invalid argument)
access("/etc/ld.so.preload", R_OK)      = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
fstat(3, {st_mode=S_IFREG|0644, st_size=12201, ...}) = 0
mmap(NULL, 12201, PROT_READ, MAP_PRIVATE, 3, 0) = 0x7f9f890d8000
close(3)                                = 0
openat(AT_FDCWD, "/lib64/librt.so.1", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\340%\0\0\0\0\0\0"..., 832) = 832
fstat(3, {st_mode=S_IFREG|0755, st_size=75272, ...}) = 0
・
・
(snip)
・
・
```

`--tt`で 行頭にタイムスタンプを付与、`-T`で行末に実行時間を付与できる。  
また、`-p`でプロセスIDを指定できる。プロセスIDをプロセス名から調べるのに、`pidof`コマンドなどもある。

```
$ tail -f /dev/null &
[1] 115
$ strace -p 115
strace: Process 115 attached
restart_syscall(<... resuming interrupted read ...>) = 0
read(3, "", 8192)                       = 0
nanosleep({tv_sec=1, tv_nsec=0}, NULL)  = 0
read(3, "", 8192)                       = 0
nanosleep({tv_sec=1, tv_nsec=0}, NULL)  = 0
read(3, "", 8192)                       = 0
nanosleep({tv_sec=1, tv_nsec=0}, NULL)  = 0
read(3, "", 8192)                       = 0
nanosleep({tv_sec=1, tv_nsec=0}, NULL)  = 0
read(3, "", 8192)                       = 0
nanosleep({tv_sec=1, tv_nsec=0}, ^Cstrace: Process 115 detached
 <detached ...>
 ```
