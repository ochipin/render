package nocache

import (
	"fmt"
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
func MakeRender(max int64, binary bool) core.Render {
	return CreateRender(&common.Config{
		Directory: "test",
		Targets:   nil,
		Exclude:   nil,
		Binary:    binary,
		MaxSize:   max,
	})
}

func MakeRenderExt(max int64, binary bool) core.Render {
	return CreateRender(&common.Config{
		Directory: "test",
		Targets:   []string{".text"},
		Exclude:   nil,
		Binary:    binary,
		MaxSize:   max,
	})
}

func Test__RENDER_CASE1(t *testing.T) {
	Render := MakeRender(1024, true)
	buf, err := Render.Render("case1/index.html", nil)
	if err != nil {
		t.Fatal("ERROR")
	}
	fmt.Println(string(buf))
}

func Test__RENDER_CASE2(t *testing.T) {
	Render := MakeRender(1024, true)
	buf, err := Render.Render("case2/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err != nil {
		t.Fatal("ERROR", err)
	}
	fmt.Println(string(buf))
}

func Test__RENDER_CASE3(t *testing.T) {
	Render := MakeRender(1024, true)
	// ファイルはあるが、case3/index.html が使用している load.html は存在しない場合
	_, err := Render.Render("case3/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)

	// 存在しないファイルを指定
	_, err = Render.Render("case3/load.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err.(*core.TemplateError).Error())

	// バイナリファイルを指定
	_, err = Render.Render("case3/binary.png", map[string]interface{}{
		"name": "Hello World",
	})
	if err != nil {
		t.Fatal("ERROR")
	}
	// バイナリファイルは認めるが、指定したサイズを超過した場合はエラーとなる
	Render = MakeRender(200, true)
	_, err = Render.Render("case3/binary.png", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	// バイナリファイルを認めない場合は、エラーとなる
	Render = MakeRender(1024, false)
	_, err = Render.Render("case3/binary.png", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}

	// 指定した拡張子以外でのアクセスは禁止
	Render = MakeRenderExt(1024, true)
	_, err = Render.Render("case3/binary.png", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
}

func Test__RENDER_CASE4(t *testing.T) {
	Render := MakeRender(1024, true)
	// パースエラー
	_, err := Render.Render("case4/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)
}

func Test__RENDER_CASE5(t *testing.T) {
	Render := MakeRender(1024, true)
	// ロード先でパースエラー
	_, err := Render.Render("case5/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)
}
func Test__RENDER_CASE6(t *testing.T) {
	Render := MakeRender(1024, true)
	// 無限ループに陥る場合
	_, err := Render.Render("case6/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)

	// ロード先でパースエラー
	_, err = Render.Render("case6/binary.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)
}

func Test__RENDER_CASE7(t *testing.T) {
	Render := MakeRender(1024, true)
	buf, err := Render.Render("case7/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err != nil {
		t.Fatal("ERROR")
	}
	fmt.Println(string(buf))
}

func Test__RENDER_CASE8(t *testing.T) {
	Render := MakeRender(1024, true)
	// バイナルデータのため、エラーとなる
	_, err := Render.Render("case8/index.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println("TEST", err)

	_, err = Render.Render("case8/parseerror.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)

	_, err = Render.Render("case8/readerror.html", map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(err)
}

func Test_HELPER_TEST(t *testing.T) {
	r1 := MakeRender(1024, true)
	if err := r1.LargeHelper(Helper1{}); err != nil {
		t.Fatal("Error")
	}
	r2 := r1.Copy()
	if err := r2.LargeHelper(Helper2{}); err != nil {
		t.Fatal("Error")
	}

	buf1, _ := r1.Render("case9/index.html", nil)
	buf2, _ := r2.Render("case9/index.html", nil)
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
	if r1.HasHelper("status") != true {
		t.Fatal("Error")
	}
	if r1.HasHelper("name") != true {
		t.Fatal("Error")
	}
}

func Test__RENDER_STRING(t *testing.T) {
	Render := MakeRender(1024, true)
	// 成功
	buf, err := Render.RenderString(`<!DOCTYPE html>
	<html>
	  <head>
		<meta charset='UTF-8' />
		<title>Test Case2</title>
	  </head>
	  <body>
		<h1>Test Case2</h1>
		{{template "case2/load.html" .}}
	  </body>
	</html>`, map[string]interface{}{
		"name": "Hello World",
	})
	if err != nil {
		t.Fatal("ERROR")
	}
	fmt.Println(string(buf))

	// 失敗
	buf, err = Render.RenderString(`<!DOCTYPE html>
	<html>
		<head>
		<meta charset='UTF-8' />
		<title>Test Case2</title>
		</head>
		<body>
		<h1>Test Case2</h1>
		{{template "case2/load.html" .}
		</body>
	</html>`, map[string]interface{}{
		"name": "Hello World",
	})
	if err == nil {
		t.Fatal("ERROR")
	}
	fmt.Println(string(buf))
}
