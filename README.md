レンダーライブラリ
===
Golangのテンプレートファイル、テンプレート文字列を処理するライブラリ。

```go
package main

import (
    "fmt"
    "regexp"

    "github.com/ochipin/render"
)

func main() {
    // テンプレートの設定
    conf := &render.Config {
        Directory:  "app/views/contents",
        Targets:    []string{".html", ".text"},
        Exclude:    regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`),
        Binary:     true,
        Cache:      true,
        MaxSize:    100 << 20,
        SumMaxSize: 1 << 30,
    }
    // テンプレート設定から、レンダーオブジェクトを生成
    c, err := conf.New()
    if err != nil {
        panic(err)
    }
    // 登録されているレンダーファイルの内容を取得
    buf, err := c.Render("app/index.html", nil)
    // 表示
    fmt.Println(string(buf))
}
```

## レンダー設定
レンダー設定用構造体、`Config`について説明する。

### Config.Directory
レンダー対象となるファイルが置かれているディレクトリを指定する。

対象となるディレクトリが、次の様な構成になっていたとする。

```
app/views/contents
  +-- dir1
  |    `-- file.html  // <h1>SAMPLE1</h1>
  `-- dir2
       +-- file.html  // <h1>SAMPLE2</h1>
       `-- file.text  // SAMPLE3
```

`Directory`に指定したディレクトリパスが、`app/views/contents` になっていた場合、
`Render`メソッドから、次のようにしてレンダーファイルの内容を取得可能。

```go
conf := &Config{Directory:"app/views/contents", ...}
c, _ := conf.New()
c.Render("dir1/file.html", nil) // <h1>SAMPLE1</h1> を取得
c.Render("dir2/file.html", nil) // <h1>SAMPLE2</h1> を取得
c.Render("dir2/file.text", nil) // SAMPLE3 を取得
```

### Config.Targets
レンダー対象となるファイルの拡張子を指定する。
指定した拡張子のみが、レンダーの対象となる。

```go
conf := &Config{
    ...
    // 指定する拡張子は、[]string型となっているため、複数指定可能
    Targets: []string{".text"},
}
```

```
app/views/contents
  +-- dir1
  |    `-- file.html
  `-- dir2
       +-- file.html
       `-- file.text  <-- dir2/file.text のみがレンダー対象となる
```
```go
c, _ := conf.New()
c.Render("dir1/file.html", nil) // error
c.Render("dir2/file.html", nil) // error
c.Render("dir2/file.text", nil) // OK
```
`Targets`パラメータになにも指定しなかった場合は、全ファイル対象となる。

### Config.Exclude
レンダー処理後に、不要となる文字列を削除する。不要となる文字列は、正規表現で指定する。

例として、次のようなJavaScriptファイルを処理するサンプルを提示する。

```js
/*=
var name = "{{.Name}}";
var type = "{{.Type}}";
 */
/*= {{template "main.js" .}} */
//= {{template "func.js" .}}
```

上記JavaScriptの、 `/*=`,`*/`,`//` を除外し、下記のようなコードに変換する。

```js
var name = "name";
var type = "type";

function main() {
    execute(name, type);
}

function execute(n, t) {
    alert(name);
    alert(type);
}
```
このようなコード変換を行う場合、`Exclude`パラメータは、次のように指定する。

```go
conf := &Config {
    // ()で囲まれた部分のみが残り、それ以外の文字列は削除される
    //                            $1             $2        $3              $4
    Exclude: regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`)
}
```

### Config.Binary
バイナリファイル取り扱いフラグ。`true`に設定することで、レンダー対象ディレクトリ内にあるバイナリファイルも、レンダー対象として取り扱う。

```go
conf := &Config {
    // 取り扱うバイナリイメージの拡張子をTargetsに追加
    Targets: []string {".html", ".text", ".png", ".jpg", ".svg"},
    Binary:  true,
}
```

```
app/views/contents
  +-- dir1
  |    `-- file.html
  +-- dir2
  |    +-- file.html
  |    `-- file.text
  `-- dir3
       +-- image.png
       +-- image.jpg
       `-- image.svg
```

```go
c, _ := conf.New()
c.Render("dir3/image.png", nil) // []byte型のimage.png情報を返却
```

### Config.Cache
キャッシュ有効無効フラグ。

* true(オンメモリ)  
高速になる反面、元データの更新等合った場合は、`New`を再実施しないと、オンメモリ上にあるファイルは更新されない。
* false(ディスク)  
低速。`Render`でレンダーファイル情報を受け取る度にディスクアクセスが生じる。

### Config.MaxSize
1つあたりのレンダーファイルの最大サイズをByte単位で指定する。指定されたサイズを超過したファイルがあった場合、`New`関数はエラーを返却する。

```go
conf := &Config {
    ...
    Targets: []string {".html", ".text", ".png", ".jpg", ".svg"},
    Binary:  true,
    // 4000 Byte を指定
    MaxSize: 4000,
}
```

```
app/views/contents
  +-- dir1
  |    `-- file.html  <-- 200B
  +-- dir2
  |    +-- file.html  <-- 210B
  |    `-- file.text  <-- 140B
  `-- dir3
       +-- image.png  <-- 1,432B
       +-- image.jpg  <-- 4,300B
       `-- image.svg  <-- 2,000B
```

```go
c, err := conf.New()
// app/views/contents/dir3/image.jpg: MaxSize(4000) < image.jpg(4300). filesize over
fmt.Println(err)
```
`MaxSize`パラメータが `"0"` 以下に設定されている場合は無制限になる。

### Config.SumMaxSize
全レンダーファイルの合計最大サイズをByte単位で指定する。指定されたサイズを超過した場合、`New`関数はエラーを返却する。

```go
conf := &Config {
    ...
    Targets:    []string {".html", ".text", ".png", ".jpg", ".svg"},
    Binary:     true,
    // 500MB を指定
    SumMaxSize: 500 << 20,
}
c, err := conf.New()
// app/views/contents: SumMaxSize(524288000) < 534388932. all filesize over
fmt.Println(err)
```
SumMaxSizeは、Cache = true の時のみ有効になる数値。

### Config.New() (Render, error)
設定した`Config`が所持する`New`関数をコールすることで、レンダー処理に使用する`Render`インタフェースを生成する。

```go
r, err := conf.New()
if err != nil {
    // エラー処理
}
r.Render("...", nil)
```

## レンダー処理
`Render`インタフェースについて説明する。

### Render(name string, data interface{}) ([]byte, error)
指定したレンダーファイルの解析結果を取得する。解析失敗時は、復帰値に error 型が返却される。
第2引数には、ビュー内で使用するデータを指定可能。

#### HTML
```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8' />
    <title>Test Case2</title>
  </head>
  <body>
    <h1>Test Case2</h1>
    <!-- Hello World を表示 -->
    {{.argsdata}}
  </body>
</html>
```

#### Render処理
```go
// テンプレートの設定
conf := &render.Config {
    // app/views/contents 配下に、app/index.html ファイルがある
    Directory:  "app/views/contents",
    Targets:    []string{".html", ".text"},
    Exclude:    regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`),
    Binary:     true,
    Cache:      true,
    MaxSize:    100 << 20,
    SumMaxSize: 1 << 30,
}
// テンプレート設定から、レンダーオブジェクトを生成
c, err := conf.New()
if err != nil {
    panic(err)
}
// 登録されているレンダーファイルの内容を取得
buf, err := c.Render("app/index.html", map[string]interface{}{
    "argsdata": "Hello World",
})
if err != nil {
    panic(err)
}
// 表示
fmt.Println(string(buf))
```

### RenderString(name string, data interface{}) ([]byte, error)
基本的には、`Render`と使用方法は同様。第一引数には、テンプレート文字列を指定することが可能。

```go
str := `<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8' />
    <title>Test Case2</title>
  </head>
  <body>
    <h1>Test Case2</h1>
    <!-- Hello World を表示 -->
    {{.argsdata}}
  </body>
</html>`
buf, err := c.RenderString(str, map[string]interface{}{
    "argsdata": "Hello World",
})
...
```

### Helper(i interface{}) error
ヘルパ関数を登録する。登録できるヘルパは、構造体型のみとなっている。登録に失敗した場合は、 error が返却される。

```go
// 独自ヘルパを作成
type MyHelper struct { ... }
func (h MyHelper) Name() string { ... }

// ヘルパを登録
r.Helper(MyHelper{})
```

登録されたヘルパは、次のようにビュー内でコールすることが可能。

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8' />
    <title>Test Case2</title>
  </head>
  <body>
    <h1>Test Case2</h1>
    <!-- MyHelper.Name -->
    {{MyHelper.Name}}
  </body>
</html>
```
`Helper`関数で登録されるメソッドは、既に登録済みのメソッドを上書きする点に、注意すること。
また、`import`, `hastemplate`という関数名は、使用出来ない点に注意すること。

### LargeHelper(i interface{}) error
使用方法は、`Helper`と同じだが、ビュー内でコールする方法が異なる。

```go
// 独自ヘルパを作成
type MyHelper struct { ... }
func (h MyHelper) Name() string { ... }

// ヘルパを登録
r.Helper(MyHelper{})
```

登録されたヘルパは、次のようにビュー内でコールすることが可能。

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8' />
    <title>Test Case2</title>
  </head>
  <body>
    <h1>Test Case2</h1>
    <!-- MyHelper.Name。 Name のみでコール可能 -->
    {{Name}}
  </body>
</html>
```

### SmallHelper(i interface{}) error
使用方法は、`Helper`と同じだが、ビュー内でコールする方法が異なる。

```go
// 独自ヘルパを作成
type MyHelper struct { ... }
func (h MyHelper) Name() string { ... }

// ヘルパを登録
r.Helper(MyHelper{})
```

登録されたヘルパは、次のようにビュー内でコールすることが可能。

```html
<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8' />
    <title>Test Case2</title>
  </head>
  <body>
    <h1>Test Case2</h1>
    <!-- MyHelper.Name。 name のみでコール可能 -->
    {{name}}
  </body>
</html>
```

### AddHelper(helper template.FuncMap) error
`template.FuncMap`ベースのヘルパを登録する。

```go
// add と min を登録
c.AddHelper(template.FuncMap{
    "add": func(a, b int) int {
        return a + b
    },
    "min": func(a, b int) int {
        return a - b
    },
})

// mul, dev を登録
c.AddHelper(template.FuncMap{
    "mul": func(a, b int) int {
        return a * b
    },
    "dev": func(a, b int) int {
        return a / b
    },
})
```

```
{{add 1 1}} {{/* 2 */}}
{{min 1 1}} {{/* 0 */}}
{{mul 2 2}} {{/* 4 */}}
{{dev 6 2}} {{/* 3 */}}
```

### HasHelper(name string) bool
指定した名前のヘルパが存在するか確認する。

```go
r.HasHelper("methodname")   // true の場合、ヘルパを所持している。
r.HasHelper("undefinename") // false の場合、ヘルパを所持していない。
```

### Copy() Render
現在のRenderをコピーする。

```go
r2 := r.Copy()
r2.Render("...", nil)
```

## import と hastemplate
ヘルパ関数名に、`「import」`、`「hastemplate」`という2つの関数名は使用できない点に注意すること。

import 関数は、`render`ライブラリが内部で実装しており、次の様な挙動をする。

```go
{{/* 何らかのテンプレートファイルのパスを $v 変数へ格納 */}}
{{$v := path}}

{{/* 変数 $v に格納されているテンプレートファイルのパスを import 関数へ渡す */}}
{{import $v}} {{/* テンプレートの解析結果を展開する */}}
```
存在しないテンプレートファイル名を指定すると、レンダリングエラーとなり、途中で処理を中断する。

`hastemplate`関数と併用することで、存在するテンプレートファイルがある場合のみ、importを実施する、という挙動にすることも可能。

```go
{{/* 何らかのテンプレートファイルのパスを $v 変数へ格納 */}}
{{$v := path}}

{{/* テンプレートファイルが存在した場合のみ、importを実施 */}}
{{if hastemplate $v}}
  {{/* 変数 $v に格納されているテンプレートファイルのパスを import 関数へ渡す */}}
  {{import $v}} <-- テンプレートの解析結果を展開する
{{end}}
```