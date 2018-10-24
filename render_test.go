package render

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type Sample struct{}

func (s Sample) HelperMethod() string {
	return "Call Helper Method"
}

// レンダリング成功テスト
func Test__RENDER_NORMAL_SUCCESS(t *testing.T) {
	// 設定情報を構築
	c := &Config{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
	}

	// 設定情報を元に、レンダリングオブジェクトを作成
	r, err := c.NewRender()
	if err != nil {
		t.Fatal(err)
	}

	cp := r.Copy()
	// データを登録
	cp.Data = map[string]interface{}{
		"title": "Test__RENDER_NORMAL_SUCCESS",
	}
	// ヘルパを登録
	if err := cp.GlobalHelper(Sample{}); err != nil {
		t.Fatal(err)
	}
	if err := cp.Helper(Sample{}); err != nil {
		t.Fatal(err)
	}
	cp.Funcs["mapfunc"] = func() string {
		return "mapfunc called"
	}
	// レンダリング開始
	buf, err := cp.Render("app/index.html")
	if err != nil {
		t.Fatal(err)
	}

	// 表示
	fmt.Println(string(buf))

	// 当然、Stringでの表示も可能
	buf, err = cp.String([]byte(`{{.title}} {{mapfunc}}`))
	if err != nil {
		t.Fatal(err)
	}
}

// レンダリング失敗
func Test__RENDER_NORMAL_FAILED(t *testing.T) {
	// 設定情報を構築
	c := &Config{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		Extension: []string{
			".html",
		},
	}

	// 設定情報を元に、レンダリングオブジェクトを作成
	r, err := c.NewRender()
	if err != nil {
		t.Fatal(err)
	}

	cp := r.Copy()
	// データを登録
	cp.Data = map[string]interface{}{
		"title": "Test__RENDER_NORMAL_SUCCESS",
	}
	// ヘルパを登録
	if err := cp.GlobalHelper(Sample{}); err != nil {
		t.Fatal(err)
	}
	if err := cp.Helper(Sample{}); err != nil {
		t.Fatal(err)
	}
	// レンダリング開始
	_, err = cp.Render("app/index.html")
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err.Error())
	}

	cp.Funcs["mapfunc"] = func() string {
		return "mapfunc called"
	}
	// 再帰的なコールの場合、エラーとなる
	_, err = cp.Render("app/recursive.html")
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err.Error())
	}

	// 構文エラー
	_, err = cp.String([]byte(`{{template "app/recursive.html" .}`))
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err.Error())
	}

	// 再起エラー
	_, err = cp.String([]byte(`{{template "app/recursive.html" .}}`))
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println(err.Error())
	}
}

func Test__RENDER_STRING_SUCCESS(t *testing.T) {
	// 文字列指定のレンダー処理
	buf, err := String([]byte(`{{.title}}`), map[string]interface{}{
		"title": "RENDER_STRING_SUCCESS",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	// RENDER_STRING_SUCCESS を表示する
	fmt.Println(string(buf))

	// 誤ったテンプレートファイルの場合、エラーとなる
	buf, err = String([]byte(`{{.title}`), map[string]interface{}{
		"title": "RENDER_STRING_SUCCESS",
	})
	if err == nil {
		t.Fatal("Failed")
	}
	fmt.Println(err.Error())

	// 未定義の関数をコールした場合、エラーとなる
	buf, err = String([]byte(`{{.title}} {{method}}`), map[string]interface{}{
		"title": "RENDER_STRING_SUCCESS",
	})
	if err == nil {
		t.Fatal("Failed")
	}
	fmt.Println(err.Error())

	buf, err = String([]byte(`{{template "string" .}}`), map[string]interface{}{
		"title": "RENDER_STRING_SUCCESS",
	})
	if err == nil {
		t.Fatal("Failed")
	}
	fmt.Println(err.Error())
}

func Test__HELPER_FAILED(t *testing.T) {
	// 設定情報を構築
	c := &Config{
		TargetDirs: []string{
			"test/content",
			"test/layout",
		},
		// 全ファイル対象とする
		Extension: []string{},
	}
	// 設定情報を元に、レンダリングオブジェクトを作成
	r, err := c.NewRender()
	if err != nil {
		t.Fatal(err.Error())
	}

	r.Funcs = nil
	if err := r.GlobalHelper(Sample{}); err != nil {
		t.Fatal(err.Error())
	}
	// 多重登録はエラーとなる
	if err := r.GlobalHelper(Sample{}); err == nil {
		t.Fatal("Failed")
	} else {
		fmt.Println(err.Error())
	}

	r.Funcs = nil
	if err := r.Helper(Sample{}); err != nil {
		t.Fatal(err.Error())
	}
	// 多重登録はエラーとなる
	if err := r.Helper(Sample{}); err == nil {
		t.Fatal("Failed")
	} else {
		fmt.Println(err.Error())
	}
	// 構造体ではない場合、エラーとなる
	if err := r.GlobalHelper(&Sample{}); err == nil {
		t.Fatal("Failed")
	} else {
		fmt.Println(err.Error())
	}
	if err := r.Helper(200); err == nil {
		t.Fatal("Failed")
	} else {
		fmt.Println(err.Error())
	}

	// コピー時
	cp := r.Copy()
	if _, ok := cp.Funcs["Sample"]; !ok {
		t.Fatal("failed")
	}
}

func Test__TEMPLATE_FUNC_FAILED(t *testing.T) {
	// 設定情報を構築
	c := &Config{
		// 対象ファイルディレクトリがない
		TargetDirs: []string{},
		// 全ファイル対象とする
		Extension: []string{},
	}
	// 設定情報を元に、レンダリングオブジェクトを作成
	r, err := c.NewRender()
	if err != nil {
		t.Fatal(err.Error())
	}

	if _, err := r.String([]byte(`OK!`)); err == nil {
		t.Fatal("Failed")
	} else {
		fmt.Println(err.Error())
	}
}

func Test__TEMPLATE_FUNC_FAILED2(t *testing.T) {
	// 設定情報を構築
	c := &Config{
		// 対象ファイルディレクトリがない
		TargetDirs: []string{
			"test/auth",
		},
		// 全ファイル対象とする
		Extension: []string{},
	}
	// テストように auth/auth.html を作成する
	os.Mkdir("test/auth", 0755)
	ioutil.WriteFile("test/auth/auth.html", []byte(""), 0)
	os.Chmod("test/auth/auth.html", 0001)

	// 設定情報を元に、レンダリングオブジェクトを作成
	_, err := c.NewRender()
	if err == nil {
		t.Fatal(err.Error())
	}

	os.Remove("test/auth/auth.html")
}
