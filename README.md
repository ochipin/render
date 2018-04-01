レンダーライブラリ
================================================================================
Golangのテンプレートファイル、テンプレート文字列を処理するライブラリです。

```go
package main

import (
    "fmt"
    "text/template"

    "github.com/ochipin/render"
)

/*
 views
   +-- content
   |      +-- app
   |           +-- index.html  // {{template "header/header.html" .}}
   |                           // index{{.extension}} {{hello}}
   |                           // {{template "footer/footer.html" .}}
   +-- _layout
          +-- header
          |    +-- header.html // header.html
          +-- footer
               +-- footer.html // footer.html
 */
func main() {
    r := &render.Render{
        TargetDirs: []string{
            "views/content", // 処理するテンプレートファイル置き場その1
            "views/_layout", // 処理するテンプレートファイル置き場その2
        },
        Extension: []string{
            ".html", ".htm",
        },
        Data: map[string]string{
            "extension": ".html",
        },
        Funcs: template.FuncMap{
            "hello": func() string{
                return "Hello World"
            },
        },
    }

    buf, err := r.Render("app/index.html")
    if err != nil {
        panic(err)
    }

    // header.html
    // index.html Hello World
    // footer.html
    fmt.Println(string(buf))
}
```