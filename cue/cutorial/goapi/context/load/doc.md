build.Instance

ImortPath => そのインスタンスを importするための unique な Path ?
Imports => そのインスタンスが import している instance のリスト？

単純に  load.Instances した場合、

Root, Dir => load したファイルの場所のディレクトリ
Module => 空文字列
pkgName => "package" で指定された package名
ImportPath => 空文字列
Imports: => nil
ImportPaths => load したファイルが import しているパッケージパス
ImportPos => ImportPaths に

```
=== bis ===
build.Instance{ctxt:(*build.Context)(0xc00031cdb0), BuildFiles:[]*build.File{(*build.File)(0xc0002f21c0)}, IgnoredFiles:[]*build.File(nil), OrphanedFiles:[]*build.File(nil), InvalidFiles:[]*build.File(nil), UnknownFiles:[]*build.File(nil), User:true, Files:[]*ast.File{(*ast.File)(0xc0002066c0)}, loadFunc:(build.LoadFunc)(0x14edc20), done:true, PkgName:"hello", hasName:true, ImportPath:"", Imports:[]*build.Instance(nil), Err:errors.Error(nil), parent:(*build.Instance)(nil), DisplayPath:"command-line-arguments", Module:"", Root:"/Users/hsugino/ghq/github.com/ginoh/sandbox/cue/cutorial/goapi/context/load", Dir:"/Users/hsugino/ghq/github.com/ginoh/sandbox/cue/cutorial/goapi/context/load", ImportComment:"", AllTags:[]string(nil), Incomplete:false, ImportPaths:[]string{"strings"}, ImportPos:map[string][]token.Pos{"strings":[]token.Pos{token.Pos{file:(*token.File)(0xc0002f22a0), offset:420}}}, Deps:[]string{}, DepsErrors:[]error(nil), Match:[]string(nil)}
Error during build: imported and not used: "strings"
```

cue.mod init example.cue.com を実行
load する際の指定をファイルじゃなくて、package指定にしてみたところ "example.cue.com:hello"

```
=== bis ===
build.Instance{ctxt:(*build.Context)(0xc00036fef0), BuildFiles:[]*build.File{(*build.File)(0xc00014bb90)}, IgnoredFiles:[]*build.File(nil), OrphanedFiles:[]*build.File{(*build.File)(0xc0001330a0)}, InvalidFiles:[]*build.File(nil), UnknownFiles:[]*build.File{(*build.File)(0xc00050c310)}, User:false, Files:[]*ast.File{(*ast.File)(0xc000285ce0)}, loadFunc:(build.LoadFunc)(0x14edc20), done:true, PkgName:"hello", hasName:true, ImportPath:"example.cue.com:hello", Imports:[]*build.Instance(nil), Err:errors.Error(nil), parent:(*build.Instance)(nil), DisplayPath:"example.cue.com:hello", Module:"example.cue.com", Root:"/Users/hsugino/ghq/github.com/ginoh/sandbox/cue/cutorial/goapi/context/load", Dir:"/Users/hsugino/ghq/github.com/ginoh/sandbox/cue/cutorial/goapi/context/load", ImportComment:"", AllTags:[]string(nil), Incomplete:false, ImportPaths:[]string{"strings"}, ImportPos:map[string][]token.Pos{"strings":[]token.Pos{token.Pos{file:(*token.File)(0xc0001331f0), offset:420}}}, Deps:[]string{}, DepsErrors:[]error(nil), Match:[]string(nil)}
Error during build: imported and not used: "strings"
```
