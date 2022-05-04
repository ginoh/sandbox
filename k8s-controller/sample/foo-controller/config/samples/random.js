// ref：https://qiita.com/ryounagaoka/items/4736c225bdd86a74d59c

// 生成する文字列の長さ
var l = 241;

// 生成する文字列に含める文字セット
var c = "abcdefghijklmnopqrstuvwxyz0123456789";

var cl = c.length;
var r = "";
for(var i=0; i<l; i++){
  r += c[Math.floor(Math.random()*cl)];
}
console.log(r)