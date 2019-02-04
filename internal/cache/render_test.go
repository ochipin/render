package cache

import (
	"regexp"
	"testing"
	"text/template"

	"github.com/ochipin/render/core"
	"github.com/ochipin/render/internal/common"
)

type Helper1 struct{}

func (h Helper1) Name() string { return "Helper1" }

type Helper2 struct{}

func (h Helper2) Name() string { return "Helper2" }

func StructHelper() core.Render {
	return CreateRender(&common.Config{
		Directory: "test",
		Files: []*common.File{
			&common.File{
				IsBinary: false,
				FileName: "app/index.html",
				FileData: []byte(`<p>{{Name}}</p>`),
			},
		},
	})
}

func CreateRenderErrors() core.Render {
	return CreateRender(&common.Config{
		Directory: "test",
		Files: []*common.File{
			&common.File{
				IsBinary: false,
				FileName: "app/errors.html",
				FileData: []byte(`<p>{{*}}</p>`),
			},
		},
	})
}

func CreateRenderSample() core.Render {
	return CreateRender(&common.Config{
		Directory: "test",
		Exclude:   regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`),
		Files: []*common.File{
			&common.File{
				IsBinary: false,
				FileName: "app/index.html",
				FileData: []byte(`<!DOCTYPE html>
<html>
  <head>
	<meta charset='UTF-8' />
	<title>app/index.html</title>
  </head>
  <body>
    {{if hastemplate "app/tmpl.html"}}{{import "app/tmpl.html"}}{{end}}
  </body>
</html>`),
			},
			&common.File{
				IsBinary: false,
				FileName: "app/tmpl.html",
				FileData: []byte(`<p>top app/index.html. app.tmpl.html template</p>`),
			},
			&common.File{
				IsBinary: true,
				FileName: "images/name.png",
				FileData: []byte(`<ping file>`),
			},
		},
	})
}

func Test_RENDER(t *testing.T) {
	r := CreateRenderSample()
	// 登録しているレンダーを取得
	_, err := r.Render("app/index.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	// 未登録のレンダーはエラーとなる
	_, err = r.Render("app/index.text", nil)
	if err == nil {
		t.Fatal("Error")
	}
	// バイナリファイルを取得
	buf, err := r.Render("images/name.png", nil)
	if string(buf) != "<ping file>" {
		t.Fatal("Error")
	}
	// 文字列テンプレートを実施する
	buf, err = r.RenderString("", nil)
	if string(buf) != "" {
		t.Fatal("Error")
	}
	// 構文エラー
	_, err = r.RenderString("{{", nil)
	if err == nil {
		t.Fatal("Error")
	}
	// 構文エラーがあるテンプレートの場合はエラーとなる
	r = CreateRenderErrors()
	_, err = r.Render("app/errors.html", nil)
	if err == nil {
		t.Fatal(err)
	}
	_, err = r.RenderString("", nil)
	if err == nil {
		t.Fatal(err)
	}

}

func Test_HELPER_TEST(t *testing.T) {
	r1 := StructHelper()
	if err := r1.LargeHelper(Helper1{}); err != nil {
		t.Fatal("Error")
	}
	r2 := r1.Copy()
	if err := r2.LargeHelper(Helper2{}); err != nil {
		t.Fatal("Error")
	}

	buf1, _ := r1.Render("app/index.html", nil)
	buf2, _ := r2.Render("app/index.html", nil)
	if string(buf1) == string(buf2) {
		t.Fatal("Error")
	}

	if err := r1.SmallHelper(Helper1{}); err != nil {
		t.Fatal("Error")
	}
	if err := r2.Helper(Helper2{}); err != nil {
		t.Fatal("Error")
	}

	r1.AddHelper(template.FuncMap{
		"status": func() string { return "status" },
	})

	if r1.HasHelper("name") != true {
		t.Fatal("Error")
	}
	if r1.HasHelper("status") != true {
		t.Fatal("Error")
	}
}
