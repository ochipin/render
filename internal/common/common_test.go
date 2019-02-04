package common

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"
)

type HelperErrors1 struct{}

// 2つめの復帰値の型が、error型ではないためエラー
func (h HelperErrors1) Case1() (int, int) {
	return 0, 0
}

type HelperErrors2 struct{}

// 復帰値が、ないためエラー
func (h HelperErrors2) Case2() {
	return
}

type HelperTest struct {
	str string
}

func (h HelperTest) TestCase1() string {
	return h.str
}

// 登録したヘルパ関数が実行可であるかをチェックする
func isSuccessHelperStruct(funcs template.FuncMap) (string, error) {
	// 登録した関数で、テンプレートを解析
	tmpl, err := template.New("test").Funcs(funcs).Parse(`{{HelperTest.TestCase1}}`)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// 登録したヘルパ関数が実行可であるかをチェックする
func isSuccessHelperLarge(funcs template.FuncMap) (string, error) {
	// 登録した関数で、テンプレートを解析
	tmpl, err := template.New("test").Funcs(funcs).Parse(`{{TestCase1}}`)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// 登録したヘルパ関数が実行可であるかをチェックする
func isSuccessHelperSmall(funcs template.FuncMap) (string, error) {
	// 登録した関数で、テンプレートを解析
	tmpl, err := template.New("test").Funcs(funcs).Parse(`{{testcase1}}`)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// 拡張子テスト
func Test_IS_SUFFIX(t *testing.T) {
	// target 未指定の場合は、全ファイルを許可する
	if HasSuffix("a.html", nil) != true {
		t.Fatal("Error")
	}
	if HasSuffix("a.text", nil) != true {
		t.Fatal("Error")
	}
	if HasSuffix("a.ping", nil) != true {
		t.Fatal("Error")
	}
	// target指定ありの場合は、指定した拡張子のもののみ許可する
	if HasSuffix("a.text", []string{".html", ".text"}) != true {
		t.Fatal("Error")
	}
	if HasSuffix("a.html", []string{".html", ".text"}) != true {
		t.Fatal("Error")
	}
	if HasSuffix("a.png", []string{".html", ".text"}) != false {
		t.Fatal("Error")
	}
}

// Windows 用のパスチェック
func Test_PATH_TO_CHECK(t *testing.T) {
	if ToWindowsPath("path\\to\\url.html") != "path/to/url.html" {
		t.Fatal("ToWindowsPath")
	}
}

// ヘルパ登録成功例のテスト
func Test_HELPER_SUCCESS_CASE(t *testing.T) {
	funcs := make(template.FuncMap)
	// 構造体をヘルパに登録
	if err := Helpers(funcs, HelperTest{str: "Sample"}, HelperStruct); err != nil {
		t.Fatal("Helpers Function Error")
	}
	result, err := isSuccessHelperStruct(funcs)
	if err != nil || result != "Sample" {
		t.Fatal("Helpers Function Error")
	}

	// 構造体をヘルパに登録
	if err := Helpers(funcs, &HelperTest{str: "Sample"}, HelperStruct); err != nil {
		t.Fatal("Helpers Function Error")
	}
	result, err = isSuccessHelperStruct(funcs)
	if err != nil || result != "Sample" {
		t.Fatal("Helpers Function Error")
	}

	// 構造体をヘルパに登録
	if err := Helpers(funcs, HelperTest{str: "Sample"}, HelperLarge); err != nil {
		t.Fatal("Helpers Function Error")
	}
	result, err = isSuccessHelperLarge(funcs)
	if err != nil || result != "Sample" {
		t.Fatal("Helpers Function Error")
	}

	// 構造体をヘルパに登録
	if err := Helpers(funcs, &HelperTest{str: "Sample"}, HelperLarge); err != nil {
		t.Fatal("Helpers Function Error")
	}
	result, err = isSuccessHelperLarge(funcs)
	if err != nil || result != "Sample" {
		t.Fatal("Helpers Function Error")
	}

	// 構造体をヘルパに登録
	if err := Helpers(funcs, HelperTest{str: "Sample"}, HelperSmall); err != nil {
		t.Fatal("Helpers Function Error")
	}
	result, err = isSuccessHelperSmall(funcs)
	if err != nil || result != "Sample" {
		t.Fatal("Helpers Function Error")
	}

	// 構造体をヘルパに登録
	if err := Helpers(funcs, &HelperTest{str: "Sample"}, HelperSmall); err != nil {
		t.Fatal("Helpers Function Error")
	}
	result, err = isSuccessHelperSmall(funcs)
	if err != nil || result != "Sample" {
		t.Fatal("Helpers Function Error")
	}
}

// ヘルパ登録失敗例
func Test_HELPER_ERROR_CASE(t *testing.T) {
	funcs := make(template.FuncMap)

	// nilをヘルパに登録
	if err := Helpers(funcs, nil, HelperSmall); err == nil {
		t.Fatal("Helpers Function Error")
	}
	// 型付きのnilをヘルパに登録
	var Nil *HelperTest
	if err := Helpers(funcs, Nil, HelperSmall); err == nil {
		t.Fatal("Helpers Function Error")
	}
	// 構造体ではないポインタ型を渡す
	var num = 200
	if err := Helpers(funcs, &num, HelperSmall); err == nil {
		t.Fatal("Helpers Function Error")
	}
	// 構造体、ポインタどちらでもない値を渡す
	if err := Helpers(funcs, num, HelperSmall); err == nil {
		t.Fatal("Helpers Function Error")
	}
	// 構造体ではあるが、メソッドが存在しない
	if err := Helpers(funcs, struct{}{}, HelperSmall); err != nil {
		t.Fatal("Helpers Function Error")
	}

	Helpers(funcs, HelperErrors1{}, HelperSmall)
	Helpers(funcs, HelperErrors2{}, HelperSmall)
}

// ヘルパ登録の名前が不正な場合、エラーとしてみなす
func Test_HELPER_METHOD_NAME(t *testing.T) {
	// ヘルパ関数未登録時は、err 値は帰ってこないはず
	if _, err := CheckFuncName(nil); err != nil {
		t.Fatal("Error")
	}
	// ヘルパ関数が認める関数名でエラーがおきないかチェック
	funcs := template.FuncMap{
		"name": func() string {
			return "name"
		},
	}
	if _, err := CheckFuncName(funcs); err != nil {
		t.Fatal("Error")
	}
	// ヘルパ関数が認めない関数名でエラーが起きるかチェック
	funcs = template.FuncMap{
		"import": func() string {
			return "name"
		},
	}
	if _, err := CheckFuncName(funcs); err == nil {
		t.Fatal("Error")
	}
	// ヘルパ関数が認めない関数名でエラーが起きるかチェック
	funcs = template.FuncMap{
		"name.args": func() string {
			return "name"
		},
	}
	if _, err := CheckFuncName(funcs); err == nil {
		t.Fatal("Error")
	}
}

func Test_ADD_HELPERS(t *testing.T) {
	var basefuncs = make(template.FuncMap)
	var addfuncs = template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"min": func(a, b int) int {
			return a - b
		},
	}

	// basefuncs に addfuncs を追加する
	FuncMapHelper(addfuncs, basefuncs)

	// basefuncs に addfuncs が登録されたか確認する
	if _, ok := basefuncs["add"]; !ok {
		t.Fatal("Error")
	}
	if _, ok := basefuncs["min"]; !ok {
		t.Fatal("Error")
	}

	// addfuncs に nil が含まれていた場合、エラーとなる
	addfuncs["name"] = nil
	if err := FuncMapHelper(addfuncs, basefuncs); err == nil {
		t.Fatal("Err")
	}

	// addfuncs に登録されている値が関数ではない場合、エラーとなる
	addfuncs["name"] = "string"
	if err := FuncMapHelper(addfuncs, basefuncs); err == nil {
		t.Fatal("Err")
	}

	// addfuncs に登録されている関数が不正な場合、エラーとなる
	addfuncs["name"] = func() {}
	if err := FuncMapHelper(addfuncs, basefuncs); err == nil {
		t.Fatal("Err")
	}

	// addfuncs に登録されている関数が不正な場合、エラーとなる
	addfuncs["name"] = func() (int, int) {
		return 0, 0
	}
	if err := FuncMapHelper(addfuncs, basefuncs); err == nil {
		t.Fatal("Err")
	}
}

func StringTemplate1(data interface{}) *template.Template {
	var tmpl *template.Template
	funcs := make(template.FuncMap)
	// import : format で指定したテンプレート名を元に、テンプレートファイルの内容をロードする
	funcs["import"] = func(format string, i ...interface{}) (string, error) {
		return Import(tmpl, data, format, i...)
	}
	// hastemplate : 指定したテンプレート名が存在するかチェックする
	funcs["hastemplate"] = func(format string, i ...interface{}) bool {
		return HasTemplate(tmpl, format, i...)
	}
	tmpl, _ = template.New("app/sample1.html").Funcs(funcs).Parse(`{{import "%s/sample2.html" "app"}}`)
	tmpl, _ = tmpl.New("app/sample2.html").Parse(`{{if hastemplate "%s/sample3.html" "app"}}{{import "app/sample3.html"}}{{end}}`)
	tmpl, _ = tmpl.New("app/sample3.html").Parse(`//= {{.Name}} `)
	return tmpl
}

func StringTemplate2(data interface{}) *template.Template {
	var tmpl *template.Template
	funcs := make(template.FuncMap)
	// import : format で指定したテンプレート名を元に、テンプレートファイルの内容をロードする
	funcs["import"] = func(format string, i ...interface{}) (string, error) {
		return Import(tmpl, data, format, i...)
	}
	// hastemplate : 指定したテンプレート名が存在するかチェックする
	funcs["hastemplate"] = func(format string, i ...interface{}) bool {
		return HasTemplate(tmpl, format, i...)
	}
	// app/sample4.html は存在しないが、ロードを実施する
	tmpl, _ = template.New("app/sample1.html").Funcs(funcs).Parse(`{{import "%s/sample4.html" "app"}}`)
	tmpl, _ = tmpl.New("app/sample2.html").Parse(`{{if hastemplate "%s/sample3.html" "app"}}{{import "app/sample3.html"}}{{end}}`)
	tmpl, _ = tmpl.New("app/sample3.html").Parse(`//= {{.Name}} `)

	return tmpl
}

func StringTemplate3(data interface{}) *template.Template {
	var tmpl *template.Template
	funcs := make(template.FuncMap)
	// import : format で指定したテンプレート名を元に、テンプレートファイルの内容をロードする
	funcs["import"] = func(format string, i ...interface{}) (string, error) {
		return Import(tmpl, data, format, i...)
	}
	// hastemplate : 指定したテンプレート名が存在するかチェックする
	funcs["hastemplate"] = func(format string, i ...interface{}) bool {
		return HasTemplate(tmpl, format, i...)
	}
	// app/sample4.html は存在しないが、ロードを実施する
	tmpl, _ = template.New("app/sample1.html").Funcs(funcs).Parse(`{{template "app/sample4.html" .}}`)

	return tmpl
}

func Test_TEMPLATE_SUCCESS(t *testing.T) {
	var data = map[string]interface{}{
		"Name": "SAMPLE NAME",
	}
	tmpl := StringTemplate1(data)
	exclude := regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`)
	// データ app/sample1.html を表示
	buf, err := Template(tmpl, "app/sample1.html", exclude, nil)
	if err != nil {
		t.Fatal(err)
	}
	// 文字列除外設定をした結果、SAMPLE NAME であればOK！
	if string(buf) != "SAMPLE NAME" {
		t.Fatal(string(buf))
	}

	// データ app/sample1.html を表示
	buf, err = Template(tmpl, "app/sample1.html", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// 文字列除外設定なしの結果が、//= SAMPLE NAME であればOK！
	if string(buf) != "//= SAMPLE NAME " {
		t.Fatal(string(buf))
	}
}

func Test_TEMPLATE_ERROR(t *testing.T) {
	var buf bytes.Buffer

	// template: app/index.html:1: unexpected unclosed action in command
	_, err := template.New("app/index.html").Parse("{{")
	RenderError(err, nil, "{{")

	// template: app/index.html:1: function "undefined_func" not defined
	_, err = template.New("app/index.html").Parse("{{undefined_func}}")
	RenderError(err, nil, "{{undefined_func}}")

	// template: app/index.html:1: missing value for command
	_, err = template.New("app/index.html").Parse("{{}}")
	RenderError(err, nil, "{{}}")

	// template: app/index.html:1: unexpected "}" in command
	_, err = template.New("app/index.html").Parse("{{}")
	RenderError(err, nil, "{{}")

	// template: app/index.html:1:11: executing "app/index.html" at <{{template "index/ap...>: template "index/app.html" not defined
	tmpl, _ := template.New("app/index.html").Parse(`{{template "index/app.html" .}}`)
	err = tmpl.ExecuteTemplate(&buf, "app/index.html", nil)
	RenderError(err, tmpl, "")

	// template: app/index.html:1:2: executing "app/index.html" at <name>: error calling name: error message
	tmpl, _ = template.New("app/index.html").Funcs(template.FuncMap{
		"name": func() (int, error) { return 0, fmt.Errorf("error message") },
	}).Parse("{{name}}")
	err = tmpl.ExecuteTemplate(&buf, "app/index.html", nil)
	RenderError(err, tmpl, "")

	// template: app/index.html:1:2: executing "app/index.html" at <import>: error calling import: template: no template "sample2.html" associated with template "app/index.html"
	tmpl, _ = template.New("sample").Funcs(template.FuncMap{
		"import": func() (string, error) {
			var buf bytes.Buffer
			err := tmpl.ExecuteTemplate(&buf, "sample2.html", nil)
			return "", err
		},
	}).Parse(`{{template "app/index.html" .}}`)
	tmpl, _ = tmpl.New("app/index.html").Parse(`{{import}}`)
	err = tmpl.ExecuteTemplate(&buf, "sample", nil)
	// fmt.Println(tmpl.Lookup("app/index.html").Root)
	RenderError(err, tmpl, "")

	// template: sample:1:11: executing "sample" at <{{template "app/inde...>: exceeded maximum template depth (100000)
	tmpl, _ = template.New("sample").Parse(`{{template "app/index.html" .}}`)
	tmpl, _ = tmpl.New("app/index.html").Parse(`{{template "sample" .}}`)
	err = tmpl.ExecuteTemplate(&buf, "sample", nil)
	RenderError(err, tmpl, "")

	// template: no template "undefined" associated with template "app/index.html"
	tmpl, _ = template.New("app/index.html").Parse(`{{template "sample.html" .}}`)
	err = tmpl.ExecuteTemplate(&buf, "undefined", nil)
	if RenderError(err, tmpl, "").Error() != `template: no template "undefined" associated with template "app/index.html"` {
		t.Fatal("Error")
	}
}

func Test_TEMPLATE_ERROR2(t *testing.T) {
	var data = map[string]interface{}{
		"Name": "SAMPLE NAME",
	}
	tmpl := StringTemplate3(data)
	exclude := regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`)
	// データ app/sample1.html を表示
	_, err := Template(tmpl, "app/sample1.html", exclude, nil)
	if err == nil {
		t.Fatal("Error")
	}

	tmpl = StringTemplate2(data)
	exclude = regexp.MustCompile(`(^|[|\n])//=\s*(.+?)\s*$|(^|[|\n])/\*=\s*([\s\S]+?)\s*\*/`)
	// データ app/sample1.html を表示
	_, err = Template(tmpl, "app/sample1.html", exclude, nil)
	if err == nil {
		t.Fatal("Error")
	}
}

func Test_READFILE(t *testing.T) {
	// バイナリファイルを読み込み
	buf, err := ReadFile("isbinary/binary.png")
	if err != nil {
		t.Fatal("Error")
	}
	// バイナリ判定がfalseの場合はテスト失敗
	if buf.IsBinary() == false {
		t.Fatal("Error")
	}
	// 全データを取得する
	b := buf.ReadAll()
	if int64(len(b)) != buf.Size() {
		t.Fatal("Error")
	}
	buf.Close()

	// バイナリファイルを読み込み
	buf, err = ReadFile("isbinary/index.html")
	if err != nil {
		t.Fatal("Error")
	}
	// バイナリ判定がtrueの場合はテスト失敗
	if buf.IsBinary() == true {
		t.Fatal("Error")
	}
	// 全データを取得する
	b = buf.ReadAll()
	if int64(len(b)) != buf.Size() {
		t.Fatal("Error")
	}
	buf.Close()

	// 存在しないデータを指定した場合、エラーとなる
	buf, err = ReadFile("isbinary/undefined")
	if err == nil {
		t.Fatal("Error")
	}
}
