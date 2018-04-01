package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

// Error 構造体は、Renderのエラー処理に使用する
type Error struct {
	Line     int
	Type     string
	Message  string
	Basename string
	Root     string
}

// Error 関数
func (e *Error) Error() string {
	if e.Line == 0 {
		if e.Type == "" {
			return e.Message
		}
		return e.Type + ": " + e.Message
	}
	// ex) render error: app/index.html:%d: function "callfunc" not defined
	return fmt.Sprintf("%s: %s:%d: %s", e.Type, e.Basename, e.Line, e.Message)
}

// Error 構造体に値をセットする関数
func newError(err error, tmpl *template.Template, root string) error {
	messages := strings.Split(err.Error(), ": ")

	switch len(messages) {
	// ex) nil or error message
	case 0:
		return nil
	case 1:
		return &Error{Message: err.Error()}
	// ex) template: error message.
	case 2:
		if messages[0] == "template" {
			messages[0] = "render"
		}
		return &Error{
			Type:    messages[0] + " error",
			Message: messages[1],
		}
	}
	// ex) template: basename:1:11: error message.
	if messages[0] == "template" {
		messages[0] = "render"
	}
	// basename:1:11 ---> basename, 1, 11 へ分割
	values := strings.Split(messages[1], ":")
	var basename string
	var line int
	if len(values) > 1 {
		// 2つ以上に分割できた場合 basename と line を設定
		basename = values[0]
		if i, err := strconv.Atoi(values[1]); err == nil {
			line = i
		}
	} else if len(values) == 1 {
		// 1 つにしか分割できなかった場合、basenameのみ設定
		basename = values[0]
	}
	// メッセージを設定
	var message = strings.Join(messages[2:], ": ")
	if tmpl != nil {
		l := tmpl.Lookup(basename)
		if l != nil && l.Root != nil {
			root = l.Root.String()
		}
	}
	return &Error{
		Line:     line,
		Basename: basename,
		Message:  message,
		Type:     messages[0] + " error",
		Root:     root,
	}
}

// Render 構造体は、テンプレートファイルの解析に使用する
type Render struct {
	TargetDirs []string         // ビュー内で使用可能なデータを登録する
	Extension  []string         // 許可する拡張子
	Data       interface{}      // ビュー内で使用可能なデータを登録する
	Funcs      template.FuncMap // ヘルパ関数登録
}

// String は、文字列テンプレートを処理する
func String(src []byte, i interface{}) ([]byte, error) {
	tmpl, err := template.New("string").Parse(string(src))
	if err != nil {
		// 登録されていない関数がコールされた、などの際に復帰値errが返却される
		return nil, newError(err, tmpl, string(src))
	}

	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "string", i); err != nil {
		// exceeded maximum template depth (100000)
		return nil, newError(err, tmpl, string(src))
	}

	return buf.Bytes(), nil
}

// Helper を登録する
func (r *Render) Helper(i interface{}) error {
	// 型情報を取得
	types := reflect.TypeOf(i)
	// 構造体ではない場合、エラーを返却する
	if types.Kind() != reflect.Struct {
		return newError(fmt.Errorf("helper: argument type not struct"), nil, "")
	}

	// メソッドが1つ以上ある場合は、Funcsへ関数を登録する
	if types.NumMethod() > 0 {
		fv := reflect.ValueOf(i)
		if r.Funcs == nil {
			r.Funcs = make(template.FuncMap)
		}
		for i := 0; i < types.NumMethod(); i++ {
			method := types.Method(i)
			r.Funcs[method.Name] = fv.Method(i).Interface()
		}
	}
	return nil
}

// String は、文字列テンプレートを処理する
func (r *Render) String(src []byte) ([]byte, error) {
	tmpl, root, err := r.filelist(r.TargetDirs...)
	if err != nil {
		return nil, newError(err, tmpl, root)
	}
	if tmpl == nil {
		t, e := template.New("string").Funcs(r.Funcs).Parse(string(src))
		if e != nil {
			return nil, newError(e, tmpl, string(src))
		}
		tmpl = t
	} else {
		t, e := tmpl.New("string").Parse(string(src))
		if e != nil {
			return nil, newError(e, tmpl, string(src))
		}
		tmpl = t
	}
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "string", r.Data); err != nil {
		// exceeded maximum template depth (100000)
		return nil, newError(err, tmpl, "")
	}
	return buf.Bytes(), err
}

// Render は複数のターゲットを1つのテンプレートファイルにする
func (r *Render) Render(viewname string) ([]byte, error) {
	tmpl, root, err := r.filelist(r.TargetDirs...)
	if err != nil {
		return nil, newError(err, tmpl, root)
	}
	if tmpl == nil {
		return nil, newError(fmt.Errorf("template: target directory not found"), nil, "")
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, viewname, r.Data); err != nil {
		return nil, newError(err, tmpl, "")
	}
	return buf.Bytes(), nil
}

// ディレクトリ配下にある対象となるファイルを抽出し、テンプレートを生成する
func (r *Render) filelist(dirs ...string) (*template.Template, string, error) {
	var tmpl *template.Template

	// 指定されたディレクトリ分ループする
	for _, dirname := range dirs {
		var root string
		err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
			// ファイルがない、またはディレクトリの場合はスルーする
			if err != nil || info.IsDir() {
				return nil
			}
			// 対象となる拡張子がない場合は、falseを返却する
			if r.isExt(path) == false {
				return nil
			}
			// ファイルを読み込み、テンプレートを生成する
			var buf []byte
			if buf, err = ioutil.ReadFile(path); err == nil {
				basename := path[len(dirname)+1:]
				root = string(buf)
				if tmpl == nil {
					tmpl, err = template.New(r.path(basename)).Funcs(r.Funcs).Parse(string(buf))
				} else {
					tmpl, err = tmpl.New(r.path(basename)).Parse(string(buf))
				}
			}
			return err
		})
		if err != nil {
			return nil, root, err
		}
	}
	return tmpl, "", nil
}

// 拡張子の一致を確認する
func (r *Render) isExt(path string) bool {
	if len(r.Extension) == 0 {
		return true
	}
	// 対象となる拡張子があるかチェックする
	for _, ext := range r.Extension {
		if len(path) < len(ext) || path[len(path)-len(ext):] == ext {
			return true
		}
	}
	// 拡張子が一致しない場合は、falseを返却する
	return false
}

// Windows の場合、\ ---> / へ置き換える
func (r *Render) path(path string) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(path, "\\", "/", -1)
	}
	// Windows 以外の場合しか、通らない
	return path
}
