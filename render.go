package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// Error 構造体は、Renderのエラー処理に使用する
type Error struct {
	Line     int    // エラー行番号
	Type     string // エラータイプ
	Message  string // エラー内容
	Basename string // エラーが発生したビューファイル名
	Root     string // エラーが発生したビューファイルの本文
}

// Error 関数
func (e *Error) Error() string {
	if e.Line == 0 {
		return e.Message
	}
	// ex) app/index.html:11: function "callfunc" not defined
	return fmt.Sprintf("%s:%d: %s", e.Basename, e.Line, e.Message)
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

// String : 文字列テンプレートを処理する
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

// バイナリファイルのみを対象としたマップデータを作成する
func toBinary(files []*File) map[string][]byte {
	result := make(map[string][]byte)
	for _, f := range files {
		if !f.IsBinary {
			continue
		}
		result[f.Filename] = []byte(f.Template)
	}
	return result
}

// Config : 対象となるテンプレートファイルの設定情報を管理する構造体である。
type Config struct {
	TargetDirs []string // ビュー内で使用可能なデータを登録する
	Extension  []string // 許可する拡張子
	// 読み込んだビューファイル内で除外する文字列を指定する
	// Exclude で指定した文字列を除外するのは、テンプレートファイルの構文解析前に実行される点に注意すること
	Exclude *regexp.Regexp
	Cache   bool // ビューのキャッシュ有効フラグ
}

// NewRender : Render構造体を作成する関数
func (c *Config) NewRender() (*Render, error) {
	// ファイルリストを作成する
	filelist, err := c.filelist(c.TargetDirs...)
	if err != nil {
		return nil, err
	}
	// バイナリファイルリストを作成する
	binlist := toBinary(filelist)
	// テンプレート情報を返却する
	return &Render{
		filelist: filelist,
		cache:    c.Cache,
		exclude:  c.Exclude,
		dirs:     c.TargetDirs,
		ext:      c.Extension,
		binlist:  binlist,
		Funcs:    make(template.FuncMap),
	}, nil
}

// ディレクトリ配下にある対象となるファイルを抽出し、テンプレートを生成する
func (c *Config) filelist(dirs ...string) ([]*File, error) {
	var filelist []*File

	// 指定されたディレクトリ分ループする
	for _, dirname := range dirs {
		err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
			// ファイルがない、またはディレクトリの場合はスルーする
			if err != nil || info.IsDir() {
				return nil
			}
			// 対象となる拡張子がない場合は、falseを返却する
			if c.isExt(path) == false {
				return nil
			}
			// ファイルを読み込み、ファイル内容を変数に格納する
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			str := string(buf)

			// ファイル内容がバイナリかテキストか判定する
			IsBinary := false
			for _, b := range buf {
				if b <= 8 {
					IsBinary = true
				}
			}

			// 文字列除外設定されている場合、指定された文字列を除外する
			if c.Exclude != nil && IsBinary == false {
				regex := c.Exclude.Copy()
				var replace = ""
				var reps = make([]string, regex.NumSubexp())
				for i := 0; i < regex.NumSubexp(); i++ {
					reps = append(reps, fmt.Sprintf("$%d", i+1))
				}
				replace = strings.Join(reps, "")
				str = regex.ReplaceAllString(str, replace)
			}

			filelist = append(filelist, &File{
				Filename: c.path(path[len(dirname)+1:]), // ファイル名
				Template: str,                           // ファイル内容
				IsBinary: IsBinary,                      // バイナリ or テキスト
			})

			return err
		})
		if err != nil {
			return nil, err
		}
	}
	// ファイルリストを返却する
	return filelist, nil
}

// 拡張子の一致を確認する
func (c *Config) isExt(path string) bool {
	if len(c.Extension) == 0 {
		return true
	}
	// 対象となる拡張子があるかチェックする
	for _, ext := range c.Extension {
		if len(path) < len(ext) || path[len(path)-len(ext):] == ext {
			return true
		}
	}
	// 拡張子が一致しない場合は、falseを返却する
	return false
}

// Windows の場合、\ ---> / へ置き換える
func (c *Config) path(path string) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(path, "\\", "/", -1)
	}
	// Windows 以外の場合しか、通らない
	return path
}

// File : テンプレート対象になるファイル情報を管理する構造体
type File struct {
	Filename string
	Template string
	IsBinary bool
}

// Render : レンダー管理構造体
type Render struct {
	mu       sync.Mutex        // ミューテックス
	filelist []*File           // テンプレートファイルリスト
	cache    bool              // テンプレートキャッシュの有効無効フラグ
	exclude  *regexp.Regexp    // 正規表現オブジェクト
	dirs     []string          // 対象ディレクトリ一覧
	ext      []string          // 対象拡張子一覧
	binlist  map[string][]byte // バイナリファイル一覧
	Funcs    template.FuncMap  // ビュー内で使用可能なヘルパ関数を登録する
	Data     interface{}       // ビュー内で使用可能なデータを登録する
}

// Copy : Render をコピーする
func (r *Render) Copy() *Render {
	var funcs = make(template.FuncMap)
	if r.Funcs != nil {
		for k, v := range r.Funcs {
			funcs[k] = v
		}
	}
	return &Render{
		filelist: r.filelist,
		exclude:  r.exclude,
		dirs:     r.dirs,
		ext:      r.ext,
		cache:    r.cache,
		binlist:  r.binlist,
		Funcs:    funcs,
		Data:     r.Data,
	}
}

// Helper : ヘルパ登録関数
func (r *Render) Helper(i interface{}) error {
	// 型情報を取得
	types := reflect.TypeOf(i)
	// 構造体ではない場合、エラーを返却する
	if types.Kind() != reflect.Struct {
		return newError(fmt.Errorf("argument type not struct"), nil, "")
	}
	// template.FuncMap が nil の場合 make する
	if r.Funcs == nil {
		r.Funcs = make(template.FuncMap)
	}
	// 重複していた場合、エラーを返却する
	if _, ok := r.Funcs[types.Name()]; ok {
		return newError(fmt.Errorf("'%s' - duplicate function", types.Name()), nil, "")
	}
	// メソッドが1つ以上ある場合は、Funcsへ関数を登録する
	if types.NumMethod() > 0 {
		r.mu.Lock()
		defer r.mu.Unlock()
		fv := reflect.New(types)
		r.Funcs[types.Name()] = func() interface{} {
			return fv.Elem().Interface()
		}
	}

	return nil
}

// GlobalHelper : ヘルパ登録関数
func (r *Render) GlobalHelper(i interface{}) error {
	// 型情報を取得
	types := reflect.TypeOf(i)

	// 構造体ではない場合、エラーを返却する
	if types.Kind() != reflect.Struct {
		return newError(fmt.Errorf("argument type not struct"), nil, "")
	}

	// メソッドが1つ以上ある場合は、Funcsへ関数を登録する
	if types.NumMethod() > 0 {
		fv := reflect.ValueOf(i)
		if r.Funcs == nil {
			r.Funcs = make(template.FuncMap)
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		for i := 0; i < types.NumMethod(); i++ {
			method := types.Method(i)
			if _, ok := r.Funcs[method.Name]; ok {
				return newError(fmt.Errorf("'%s' - duplicate function", method.Name), nil, "")
			}
			r.Funcs[method.Name] = fv.Method(i).Interface()
		}
	}
	return nil
}

// String : 文字列テンプレートを処理する
func (r *Render) String(src []byte) ([]byte, error) {
	tmpl, root, err := r.templates()
	if err != nil {
		return nil, newError(err, tmpl, root)
	}

	t, err := tmpl.New("string").Parse(string(src))
	if err != nil {
		return nil, newError(err, tmpl, string(src))
	}
	tmpl = t

	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "string", r.Data); err != nil {
		// exceeded maximum template depth (100000)
		return nil, newError(err, tmpl, "")
	}
	return buf.Bytes(), err
}

// Render : 複数テンプレートから描画する
func (r *Render) Render(viewname string) ([]byte, error) {
	// バイナリファイルを指定された場合は、バイナリデータを返却する
	if buf, err, used := r.hasBinarys(viewname); used {
		if err != nil {
			return nil, err
		}
		return buf, nil
	}

	// テキストファイルのレンダリングの場合は、テンプレートファイルを読み込む
	tmpl, root, err := r.templates()
	if err != nil {
		return nil, newError(err, tmpl, root)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, viewname, r.Data); err != nil {
		return nil, newError(err, tmpl, "")
	}
	return buf.Bytes(), nil
}

// バイナリファイルを処理する
func (r *Render) hasBinarys(viewname string) ([]byte, error, bool) {
	buf, ok := r.binlist[viewname]
	// バイナリファイルが見つからない場合、復帰する
	if !ok {
		return nil, fmt.Errorf("not found %s", viewname), false
	}
	// バイナリファイルが見つかった場合、バイナリデータを返却する
	if r.cache {
		return buf, nil, true
	}

	// キャッシュが無効の場合、ディスクを再度読み込みに行く
	conf := &Config{
		Exclude:    r.exclude,
		TargetDirs: r.dirs,
		Extension:  r.ext,
		Cache:      r.cache,
	}
	// ファイル一覧を作成
	files, err := conf.filelist(r.dirs...)
	if err != nil {
		return nil, err, true
	}
	// バイナリリストを作成する
	binlist := toBinary(files)
	// バイナリファイルが見つからない場合、復帰する
	if buf, ok = binlist[viewname]; !ok {
		return nil, fmt.Errorf("not found %s", viewname), true
	}

	return buf, nil, true
}

// テンプレートを作成する
func (r *Render) templates() (*template.Template, string, error) {
	var tmpl *template.Template
	var target string
	var err error

	// 登録されているr.Funcsに、不正な関数名が登録されていないかチェックする
	regex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]+$`)
	for name, _ := range r.Funcs {
		if regex.MatchString(name) == false {
			return nil, "", fmt.Errorf("function name %s is not a valid identifier", name)
		}
	}

	files := r.filelist
	if r.cache == false {
		conf := &Config{
			Exclude:    r.exclude,
			TargetDirs: r.dirs,
			Extension:  r.ext,
			Cache:      r.cache,
		}
		files, err = conf.filelist(r.dirs...)
		if err != nil {
			return nil, "", err
		}
	}

	// 対象となるファイル数分ループし、テンプレートを作成する
	for _, v := range files {
		// バイナリファイルの場合は、スルーする
		if v.IsBinary {
			continue
		}
		target = v.Template
		if tmpl == nil {
			tmpl, err = template.New(v.Filename).Funcs(r.Funcs).Parse(target)
		} else {
			tmpl, err = tmpl.New(v.Filename).Parse(target)
		}
		if err != nil {
			return nil, target, err
		}
	}
	// テンプレートが作成できなかった場合、エラーを返却する
	if tmpl == nil {
		return nil, "", fmt.Errorf("no target directory")
	}

	return tmpl, "", nil
}
