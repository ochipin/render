レンダーライブラリ
================================================================================
Golangのテンプレートファイル、テンプレート文字列を処理するライブラリです。

```go
package main

import (
	"fmt"

	"github.com/ochipin/render"
)

type Hello struct{}

func (h Hello) World() string {
	return "World"
}

/*
 views
   +-- content
   |      `-- app
   |           `-- index.html  // {{template "header/header.html" .}}
   |                           // index file. {{.message}} {{World}}
   |                           // {{template "footer/footer.html" .}}
   `-- _layout
          `-- header
          |    `-- header.html // header.html
          `-- footer
               `-- footer.html // footer.html
 */

func main() {
	// 設定情報を構築する
	c := &render.Config{
		TargetDirs: []string{
			"views/contents", // 処理するテンプレートファイル置き場その1
			"views/_layout",  // 処理するテンプレートファイル置き場その2
		},
		Extension: []string{
			".html", ".htm", // 対象となるファイルの拡張子
		},
		// 除外する文字列を選定可能。
		Exclude: regexp.MustCompile(`([|\n])//=\s*(.+)|([|\n])/\*=\s*([\s\S]+?)\*/`),
		// 読み込んだテンプレートファイルをキャッシュする。
		// 未指定の場合は、Render関数実行時に毎回ディスクにテンプレートファイルを読み込みに行く
		Cache: true,
	}

	// レンダーオブジェクトを生成
	r, err := c.NewRender()
	if err != nil {
		panic(err)
	}

	// ヘルパ関数を登録
	r.GlobalHelper(Hello{})
	// r.Helper(Hello{}) とすると、ビュー内では、次のようにコールすることになる。
	// {{Hello.World}}

	// ビュー内で使用するデータを登録 {{.message}} とすることでビュー内から使用可能
	r.Data = map[string]interface{}{
		"message": "Hello",
	}

	// レンダー開始
	buf, err := r.Render("app/index.html")
	if err != nil {
		panic(err)
	}
	// r.String(`This is Template {{template "app/index.html" .}}`)
	// とすることで、文字列から展開もできる

	// 既存のレンダーオブジェクトを引き継いだまま、別の用途でレンダーオブジェクトを使用したい場合は、Copy関数を使用する。
	// cp := r.Copy()
	// cp.Render("app/hello.html")

	// header.html
	// index file. Hello World
	// footer.html
	fmt.Println(string(buf))
}
```

レンダー時に失敗した場合エラーが返却される。エラーの詳細を知りたい場合は、次のようにすることで詳細を取得可能。

```go
	// レンダー開始
	buf, err := r.Render("app/index.html")
	if err != nil {
		e, _ := err.(*render.Error)
		// e.Line: 構文エラーなどが発生した際の行番号
		// e.Type: エラータイプ
		// e.Message: エラーメッセージ
		// e.Basename: エラーが発生したファイル名
		// e.Root: エラーが発生したビューファイルの本文
	}
```


## テンプレート関数 `import` と `hastemplate`

`render.Config`構造体の`Funcs`パラメータに、テンプレート内で使用する関数を定義することが可能だが、
登録する関数名に、`「import」`、`「hastemplate」`という2つの関数名は使用できない点に注意すること。

import 関数は、`render`ライブラリが内部で実装しており、次の様な挙動をする。

```go
// {{/* 何らかのテンプレートファイルのパスを $v 変数へ格納 */}}
// {{$v := path}}

// {{/* 変数 $v に格納されているテンプレートファイルのパスを import 関数へ渡す */}}
// {{import $v}} <-- テンプレートの解析結果を展開する
```
存在しないテンプレートファイル名を指定すると、レンダリングエラーとなり、途中で処理を中断する。

`hastemplate`関数と併用することで、存在するテンプレートファイルがある場合のみ、importを実施する、という挙動にすることも可能。

```go
// {{/* 何らかのテンプレートファイルのパスを $v 変数へ格納 */}}
// {{$v := path}}

// {{/* テンプレートファイルが存在した場合のみ、importを実施 */
// {{if hastemplate $v}}
//   {{/* 変数 $v に格納されているテンプレートファイルのパスを import 関数へ渡す */}}
//   {{import $v}} <-- テンプレートの解析結果を展開する
// {{end}}
```

## Exclude パラメータ

`render.Config`構造体に`Exclude`パラメータがある。
このパラメータに、除外したい文字列を設定することで、テンプレートファイル解析後に指定した文字列を除外する。

```go
// 設定情報を構築する
c := &render.Config{
    ...
    // 下記例の場合、 "//=" が先頭に記載されているコメントは除外する
    // また、 /*= から始まり、 */ で終わるコメントは除外する
    // という意味になり、() で囲まれた部分のみが残る
    // 下記で設定した正規表現の場合は、次のようにテンプレートが展開される
    Exclude: regexp.MustCompile(`([|\n])//=\s*(.+)|([|\n])/\*=\s*([\s\S]+?)\*/`),
    ...
}
```

app.css, default.css の2つのCSSファイルを例にして説明すると、次のような挙動になる。

### app.css
default.cssを読み込む。
```css
@charset "UTF-8";

/*=
{{import "default.css"}}
 */
```

### default.css
app.cssから読み込まれるCSS。
```css
html,body,pre,p,table,th,td{
  margin:0;
  padding:0;
}
...
```

### 展開後の app.css
テンプレートファイル展開後、/*= と */ が除外されている。

```css
@charset "UTF-8";

html,body,pre,p,table,th,td{
  margin:0;
  padding:0;
}
...
```

以上