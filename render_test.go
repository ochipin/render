package render

import (
	"fmt"
	"testing"
)

type Helper struct{}

func (helper Helper) Helper() string {
	return "Helper Call Function"
}

type MustHelper struct {
	Helper
	Value int
}

func (helper MustHelper) HelperMethod() string {
	helper.changeValue()
	return fmt.Sprint(helper.Value) + " " + helper.Helper.Helper()
}

func (helper *MustHelper) changeValue() {
	helper.Value = 200
}

// 正常系テスト
func TestNormalRender(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// ヘルパを登録
	if err := r.Helper(MustHelper{}); err != nil {
		t.Fatal(err)
	}

	// Funcs へ関数を登録
	r.Funcs["mapfunc"] = func() string {
		return "Render.Funcs call mapfunc()"
	}

	// レンダー開始
	buf, err := r.Render("app/index.html")
	if err != nil {
		t.Fatal(err)
	}

	// 表示
	fmt.Println(string(buf))
}

// 正常系テスト
func TestNormalRenderExt(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// ヘルパを登録
	if err := r.Helper(MustHelper{}); err != nil {
		t.Fatal(err)
	}

	// Funcs へ関数を登録
	r.Funcs["mapfunc"] = func() string {
		return "Render.Funcs call mapfunc()"
	}

	// レンダー開始
	buf, err := r.Render("app/main.htm")
	if err != nil {
		t.Fatal(err)
	}

	// 表示
	fmt.Println(string(buf))
}

// 文字列テンプレートのテスト
func TestRenderString(t *testing.T) {
	// 正常系テスト
	buf, err := String([]byte("{{.hello}} World"), map[string]string{
		"hello": "Hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(buf))

	// 異常系テスト
	if _, err := String([]byte("{{hello}} World"), nil); err == nil {
		t.Fatal("String([]byte(\"{{hello}} World\"), nil)")
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
	// 異常系テスト
	if buf, err := String(nil, nil); err != nil {
		t.Fatal("String(nil, nil)")
	} else {
		fmt.Println(len(buf))
	}
	// 異常系テスト
	if _, err := String([]byte("{{.ok World"), 200); err == nil {
		t.Fatal("String([]byte(\"{{200}} World\"), nil)")
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
	// 異常系テスト
	if _, err := String([]byte("{{template \"render.string\" .}}"), 200); err == nil {
		t.Fatal("String([]byte(\"{{200}} World\"), nil)")
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
}

func TestNormalStringRenderStruct(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// ヘルパを登録
	if err := r.Helper(MustHelper{}); err != nil {
		t.Fatal(err)
	}

	// Funcs へ関数を登録
	r.Funcs["mapfunc"] = func() string {
		return "Render.Funcs call mapfunc()"
	}

	// レンダー開始
	buf, err := r.String([]byte(`render.string = {{template "app/index.html" .}}`))
	if err != nil {
		t.Fatal(err)
	}

	// 表示
	fmt.Println(string(buf))

	buf, err = r.String([]byte(`{{template "app/recursive.html" .}}`))
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
}

// 対象となるレンダーディレクトリがnilの場合のテスト
func TestErrorTargetRender(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: nil,
		Extension: []string{
			".html",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// レンダー開始
	_, err := r.Render("app/index.html")
	if err == nil {
		t.Fatal("error")
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
}

// 存在しない関数がコールされた場合のテスト
func TestErrorCallFuncRender(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// レンダー開始
	_, err := r.Render("app/index.html")
	if err == nil {
		t.Fatal("error")
	} else {
		fmt.Println(err)
	}
}

// 存在しないパスが渡された場合のテスト
func TestErrorRender(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// ヘルパを登録
	if err := r.Helper(MustHelper{}); err != nil {
		t.Fatal(err)
	}

	// Funcs へ関数を登録
	r.Funcs["mapfunc"] = func() string {
		return "Render.Funcs call mapfunc()"
	}

	// レンダー開始
	_, err := r.Render("notfound.html")
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err)
	}
}

// ヘルパ登録時に、構造体ではない値が渡された場合のテスト
func TestErrorHelperRender(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// ヘルパを登録
	if err := r.Helper(&MustHelper{}); err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err)
	}

	// レンダー開始
	_, err := r.Render("notfound.html")
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err, err.(*Error).Line)
	}
}

func TestNormalRenderString(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: nil,
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// レンダー開始
	buf, err := r.String([]byte("hello {{.hello}}"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(buf))
}

func TestErrorRenderString(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}
	// ヘルパを登録
	if err := r.Helper(MustHelper{}); err != nil {
		t.Fatal(err)
	}

	// Funcs へ関数を登録
	r.Funcs["mapfunc"] = func() string {
		return "Render.Funcs call mapfunc()"
	}
	// レンダー開始
	_, err := r.String([]byte("hello {{hellook}}"))
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
}

func TestErrorRenderString2(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: nil,
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// レンダー開始
	_, err := r.String([]byte("hello {{hello}}"))
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err.Error())
	}
}

func TestErrorRenderString3(t *testing.T) {
	// レンダー構造体を生成
	r := Render{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Data: map[string]interface{}{
			"title": "test",
			"hello": "Hello World",
		},
	}

	// レンダー開始
	_, err := r.String([]byte("hello {{hellook}}"))
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err, err.(*Error).Root)
	}
}
