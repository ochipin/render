package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/ochipin/render/core"
)

const (
	// HelperStruct : 構造体ベースのヘルパを登録
	HelperStruct = 1
	// HelperLarge : 構造体のメソッド名でのヘルパを登録
	HelperLarge = 2
	// HelperSmall : 構造体のメソッド名を小文字に変えたヘルパを登録
	HelperSmall = 3
)

// File : 読み込んだファイルの情報を管理する構造体
type File struct {
	FileData []byte // ファイルデータ
	FileName string // ファイル名
	IsBinary bool   // バイナリデータの場合は true が格納される
}

// Config : レンダー情報の設定状況を受け取るための構造体
type Config struct {
	Directory string
	Targets   []string
	Exclude   *regexp.Regexp
	Binary    bool
	MaxSize   int64
	Files     []*File
}

// HasSuffix : 指定されている拡張子と、fnameに格納されている拡張子が一致するか確認する
func HasSuffix(fname string, targets []string) bool {
	// targets に拡張子が指定されていなければ、すべてのファイルを許可する
	if len(targets) == 0 {
		return true
	}
	// 対象となる拡張子が指定されているか確認する
	for _, ext := range targets {
		// ex) index.html(10) >= .html(5) && fname[10-5:] == ".html"
		// 拡張子が一致していたら、true を返却する
		if len(fname) >= len(ext) && fname[len(fname)-len(ext):] == ext {
			return true
		}
	}
	// 指定されている拡張子と一致しなければfalseを返却する
	return false
}

// ToWindowsPath で、path\to\url --> path/to/url へ置き換える
func ToWindowsPath(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

// Helpers : ヘルパ登録関数
func Helpers(funcs template.FuncMap, i interface{}, HelperType int) error {
	val := reflect.ValueOf(i)
	// nil が渡されている場合、エラーを返却する
	if val.IsValid() == false {
		return &core.HelperInvalid{
			Message: fmt.Sprint("helper argument is nil"),
			Type:    "",
			Kind:    "nil",
		}
	}

	name := val.Type().String()
	kind := val.Type().Kind().String()
	switch val.Kind() {
	// ポインタの場合
	case reflect.Ptr:
		// 中身を指すようにする
		for val.Kind() == reflect.Ptr {
			// ポインタが nil、もしくは Invalid の場合はエラーを返却する
			if val.IsNil() || val.IsValid() == false {
				return &core.HelperInvalid{
					Message: fmt.Sprintf("%s: helper argument is invalid", name),
					Type:    name,
					Kind:    kind,
				}
			}
			val = val.Elem()
		}
		// 構造体型ではない場合、エラーを返却する
		if val.Kind() != reflect.Struct {
			return &core.HelperInvalid{
				Message: fmt.Sprintf("%s: helper argument is not struct type", name),
				Type:    name,
				Kind:    kind,
			}
		}
		// チェックが成功した場合は、速やかに元のデータへ戻す
		val = reflect.ValueOf(i)
	// 構造体の場合、何もせず次の処理へ
	case reflect.Struct:
	// ポインタ、または構造体ではない場合、エラーを返却する
	default:
		return &core.HelperInvalid{
			Message: fmt.Sprintf("%s: helper argument is not struct type", name),
			Type:    name,
			Kind:    kind,
		}
	}

	// メソッドが1つもない場合は、ヘルパは登録しない
	if val.Type().NumMethod() == 0 {
		return nil
	}

	// メソッド名の復帰値が不正な場合、エラーを返却する
	if funcname, resnum := checkMethodResult(val.Type()); resnum != -1 {
		return &core.HelperInvalid{
			Message: fmt.Sprintf("%s: can't install method/function \"%s\" with %d results", name, funcname, resnum),
			Type:    name,
			Kind:    kind,
		}
	}

	// メソッドを登録する
	switch HelperType {
	// 構造体型として登録
	case HelperStruct:
		// Helper{} => Helper.MethodName でコール可能
		funcs[val.Type().Name()] = func() interface{} {
			return val.Interface()
		}
	// 構造体のメソッド名の大文字で登録
	case HelperLarge:
		// Helper{} => MethodName でコール可能
		for i := 0; i < val.Type().NumMethod(); i++ {
			method := val.Type().Method(i)
			funcs[method.Name] = val.Method(i).Interface()
		}
	// 構造体のメソッド名の小文字で登録
	case HelperSmall:
		// Helper{} => methodname でコール可能
		for i := 0; i < val.Type().NumMethod(); i++ {
			method := val.Type().Method(i)
			funcs[strings.ToLower(method.Name)] = val.Method(i).Interface()
		}
	}

	return nil
}

// FuncMapHelper : template.FuncMap 型でヘルパを登録する
func FuncMapHelper(addfuncs, basefuncs template.FuncMap) error {
	// これから登録する template.FuncMap 型に登録されているメソッド群をループで処理
	for k, v := range addfuncs {
		val := reflect.ValueOf(v)
		// 登録データが nil の場合、エラーを返却する
		if val.IsValid() == false {
			return &core.HelperInvalid{
				Message: fmt.Sprintf("%s is nil", k),
				Type:    "nil",
				Kind:    "nil",
			}
		}
		// 登録すべき関数が、関数型ではない場合 HelperInvalid を返却する
		if val.Kind() != reflect.Func {
			return &core.HelperInvalid{
				Message: fmt.Sprintf("%s is not function", k),
				Type:    val.Type().String(),
				Kind:    val.Kind().String(),
			}
		}
		// メソッド名の復帰値が不正な場合、エラーを返却する
		if funcname, resnum := checkMethodResult(val.Type()); resnum != -1 {
			return &core.HelperInvalid{
				Message: fmt.Sprintf("%s: can't install method/function \"%s\" with %d results", k, funcname, resnum),
				Type:    val.Type().String(),
				Kind:    val.Kind().String(),
			}
		}
		basefuncs[k] = v
	}
	return nil
}

// 登録するヘルパのメソッドの復帰値が適切かチェックする
func checkMethodResult(types reflect.Type) (string, int) {
	// 構造体が指定されている場合
	if types.Kind() == reflect.Struct {
		// 構造体に付与されているメソッド分チェックする
		for i := 0; i < types.NumMethod(); i++ {
			method := types.Method(i)
			// 復帰値が3つ以上、または0個の場合、エラーとして扱う
			if method.Type.NumOut() > 2 || method.Type.NumOut() == 0 {
				return method.Name, method.Type.NumOut()
			}
			// 復帰値が2つの場合、2つ目の復帰値が error 型ではない場合エラーとして扱う
			if method.Type.NumOut() == 2 && method.Type.Out(1).String() != "error" {
				return method.Name, method.Type.NumOut()
			}
		}
	} else if types.Kind() == reflect.Func {
		// 復帰値が3つ以上、または0個の場合、エラーとして扱う
		if types.NumOut() > 2 || types.NumOut() == 0 {
			return types.Name(), types.NumOut()
		}
		// 復帰値が2つの場合、2つ目の復帰値が error 型ではない場合エラーとして扱う
		if types.NumOut() == 2 && types.Out(1).String() != "error" {
			return types.Name(), types.NumOut()
		}
	}
	return "", -1
}

// CheckFuncName : 登録されている関数名が不正な名前で登録されていなか確認を行う
func CheckFuncName(renderfuncs template.FuncMap) (template.FuncMap, error) {
	var funcs = make(template.FuncMap)

	// 登録済みの関数名が、不正な名前になっていないかチェックする
	funcname := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	for name, fn := range renderfuncs {
		// 不正な名前で登録されていた場合は、エラーとして扱う
		if funcname.MatchString(name) == false {
			return nil, fmt.Errorf("function name %s is not a valid identifier", name)
		}
		// import, hastemplate という名前の場合は、エラーとして扱う
		if name == "import" || name == "hastemplate" {
			return nil, fmt.Errorf("'%s' function already exists", name)
		}
		funcs[name] = fn
	}
	return funcs, nil
}

// Import : 指定されたテンプレート名でテンプレートファイルを解析する
func Import(tmpl *template.Template, data interface{}, format string, i ...interface{}) (string, error) {
	// テンプレート名を変数へ格納
	var tmplname = format
	if len(i) >= 1 {
		tmplname = fmt.Sprintf(format, i...)
	}
	// テンプレート名から、該当するテンプレートファイルをロードする
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, tmplname, data); err != nil {
		return "", err
	}
	// テンプレートファイルの内容を返却する
	return buf.String(), nil
}

// HasTemplate : 指定されたテンプレート名でテンプレートファイルが存在するかチェックする
func HasTemplate(tmpl *template.Template, format string, i ...interface{}) bool {
	// テンプレート名を変数へ格納
	var tmplname = format
	if len(i) >= 1 {
		tmplname = fmt.Sprintf(format, i...)
	}
	return tmpl.Lookup(tmplname) != nil
}

// Template : テンプレート解析結果を返却する
func Template(tmpl *template.Template, tmplname string, exclude *regexp.Regexp, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	// テンプレートファイルの解析結果をバッファへ保持
	if err := tmpl.ExecuteTemplate(&buf, tmplname, data); err != nil {
		return nil, RenderError(err, tmpl, "")
	}
	return Exclude(buf.String(), exclude), nil
}

// Exclude : 指定した正規表現で文字列除外を行う
func Exclude(text string, regex *regexp.Regexp) []byte {
	// 文字列除外設定がされていない場合は、解析結果を返却する
	if regex == nil {
		return []byte(text)
	}
	var exclude = regex.Copy()
	// ()で囲まれた部分のみが、残る。それ以外は、削除される。
	//      $1             $2   $3              $4
	// ex: (^|[|\n])//=\s*(.+)|(^|[|\n])/\*=\s*([\s\S]+?)\*/
	var replaces = make([]string, exclude.NumSubexp())
	for i := 0; i < exclude.NumSubexp(); i++ {
		replaces = append(replaces, fmt.Sprintf("$%d", i+1))
	}
	replace := strings.Join(replaces, "")

	// 作成した正規表現をもとに、文字列を除外する
	for {
		if exclude.MatchString(text) == false {
			break
		}
		text = exclude.ReplaceAllString(text, replace)
	}

	// 解析結果を返却する
	return []byte(text)
}

// RenderError : レンダーエラー発生時に、エラー内容を生成する
func RenderError(err error, tmpl *template.Template, root string) error {
	var result = &core.RenderError{Message: err.Error()}

	// エラー内容を分割する
	// ex) template: app/index.html:10:28: executing ...
	fields := strings.Fields(err.Error())
	// 分割したエラー内容から、2フィールド目の app/index.html:10:28: の部分を取得する
	if len(fields) > 1 {
		// 2番目のカラムに":"が存在していない場合は、TemplateErrorとする
		if strings.Index(fields[1], ":") == -1 {
			return &core.TemplateError{Message: err.Error()}
		}
		fields = strings.Fields(strings.Replace(fields[1], ":", " ", -1))
	}

	result.Basename = fields[0]
	// 分割された [app/index.html, 10, 28] の配列をそれぞれRenderError構造体へ格納する
	if len(fields) >= 3 {
		// カラムが存在していた場合
		result.Line, _ = strconv.Atoi(fields[1])
		result.Column, _ = strconv.Atoi(fields[2])
	} else if len(fields) >= 2 {
		// カラムが存在していない場合
		result.Line, _ = strconv.Atoi(fields[1])
	}

	// エラー発生箇所を格納する
	if tmpl != nil && tmpl.Lookup(result.Basename) != nil {
		// ExecuteTemplate エラー時はtemplateを利用
		result.Root = tmpl.Lookup(result.Basename).Root.String()
	} else {
		// Parse エラー時は引数のrootを利用
		result.Root = root
	}

	// import, template による読み込みに失敗した場合
	r := regexp.MustCompile(`.+error calling .+: .+: no template \"([^"]*?)\" .+`)
	if r.MatchString(err.Error()) {
		result.Target = r.ReplaceAllString(err.Error(), "$1")
	}

	// 存在しないテンプレートファイルを指定された場合
	r = regexp.MustCompile(`.+at .+: template \"([^"]*?)\" not defined.*`)
	if r.MatchString(err.Error()) {
		result.Target = r.ReplaceAllString(err.Error(), "$1")
	}

	return result
}

// ReadFile : 指定されたファイルを読み込む
func ReadFile(fname string) (*Buf, error) {
	// 指定されたファイルを読み込む
	fp, err := os.Open(fname)
	// ファイルが存在しない、またはパーミッションがない場合はエラーとする
	if err != nil {
		return nil, err
	}
	return &Buf{fp}, nil
}

// Buf : ReadFileで指定したファイルポインタを管理する構造体
type Buf struct {
	file *os.File
}

// ReadAll : 全データを取得する
func (b *Buf) ReadAll() []byte {
	// オフセット位置を最初に戻す
	b.file.Seek(0, os.SEEK_SET)
	// 全データをバッファへコピーする
	buf := make([]byte, b.Size())
	io.ReadFull(b.file, buf)
	return buf
}

// IsBinary : バイナリデータか否かを判定する
func (b *Buf) IsBinary() bool {
	// ファイルサイズに応じて、読み込むバッファサイズを変更する
	size := b.Size()
	if size > 1024 {
		size = 1024
	}
	var buf = make([]byte, size)

	// オフセット位置を最初に戻す
	b.file.Seek(0, os.SEEK_SET)
	// ファイルの内容の一部をバッファへコピー
	reader := io.LimitReader(b.file, size)
	reader.Read(buf)

	// コピーされたデータが、バイナリファイルか否かを判定する
	for _, v := range buf {
		if v <= 8 {
			return true
		}
	}
	return false
}

// Size : ファイルサイズを返却する
func (b *Buf) Size() int64 {
	f, _ := b.file.Stat()
	return f.Size()
}

// Close : ファイルをクローズする
func (b *Buf) Close() {
	if b.file != nil {
		b.file.Close()
	}
	b.file = nil
}
