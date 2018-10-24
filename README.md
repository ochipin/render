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
   |      +-- app
   |           +-- index.html  // {{template "header/header.html" .}}
   |                           // index file. {{.message}} {{World}}
   |                           // {{template "footer/footer.html" .}}
   +-- _layout
          +-- header
          |    +-- header.html // header.html
          +-- footer
               +-- footer.html // footer.html
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

以上