package nocache

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sync"
	"text/template"

	"github.com/ochipin/render/core"
	"github.com/ochipin/render/internal/common"
)

// Template : 独自実装のテンプレート
type Template struct {
	*template.Template
	funcs template.FuncMap
}

// Render : キャッシュなしのRenderオブジェクトを管理する構造体
type Render struct {
	mu        sync.Mutex
	directory string
	targets   []string
	exclude   *regexp.Regexp
	binary    bool
	maxsize   int64
	funcs     template.FuncMap
}

// Copy : 現在のRenderをコピーする
func (r *Render) Copy() core.Render {
	// レンダーが所有しているヘルパをコピー
	var funcs = make(template.FuncMap)
	for k, v := range r.funcs {
		funcs[k] = v
	}
	// レンダーオブジェクトを返却する
	return &Render{
		directory: r.directory,
		targets:   r.targets,
		exclude:   r.exclude,
		binary:    r.binary,
		maxsize:   r.maxsize,
		funcs:     funcs,
	}
}

// HasHelper : 指定した名前のヘルパメソッドを所持しているか確認する
func (r *Render) HasHelper(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.funcs[name]
	return ok
}

// Helper : ヘルパ登録を実施する。登録できるヘルパは構造体型のみ
func (r *Render) Helper(i interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 登録されたヘルパは、構造体名.メソッド名でコール可能
	return common.Helpers(r.funcs, i, common.HelperStruct)
}

// LargeHelper : ヘルパ登録を実施する。登録できるヘルパは構造体型のみ
func (r *Render) LargeHelper(i interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 登録されたヘルパは、構造体のメソッド名のみでコール可能
	return common.Helpers(r.funcs, i, common.HelperLarge)
}

// SmallHelper : ヘルパ登録を実施する。登録できるヘルパは構造体型のみ
func (r *Render) SmallHelper(i interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 登録されたヘルパは、構造体のメソッド名のみでコール可能
	// メソッド名が、Hello の場合、呼び出す側は hello でコール
	return common.Helpers(r.funcs, i, common.HelperSmall)
}

// AddHelper : template.FuncMap そのものを登録する
func (r *Render) AddHelper(helper template.FuncMap) error {
	return common.FuncMapHelper(helper, r.funcs)
}

// RenderString : 文字列レンダーを処理する
func (r *Render) RenderString(text string, data interface{}) ([]byte, error) {
	// レンダーファイルの場合はパース開始
	tmpl, err := r.template("string", []byte(text), data)
	// パースエラーが発生した場合は、エラーを返却する
	if err != nil {
		return nil, err
	}
	// パースデータを実行する
	return r.execute(tmpl, "string", data)
}

// Render : 指定した名前でデータでテンプレートファイルの解析結果を取得する
func (r *Render) Render(tmplname string, data interface{}) ([]byte, error) {
	// 指定されたファイル名をリードする
	buf, isBinary, err := r.readfile(tmplname)
	if err != nil {
		return nil, err
	}
	// バイナリファイルの場合は、バイナリデータを返却する
	if isBinary {
		return buf, nil
	}
	// レンダーファイルの場合はパース開始
	tmpl, err := r.template(tmplname, buf, data)
	// パースエラーが発生した場合は、エラーを返却する
	if err != nil {
		return nil, err
	}

	// パースデータを実行する
	return r.execute(tmpl, tmplname, data)
}

// 指定した名前のファイルを読み込み、データを返却する。バイナリの場合は、2つの目の復帰値が true になる
func (r *Render) readfile(name string) ([]byte, bool, error) {
	// 登録済みの拡張子と一致しない場合は、エラーを返却する
	if common.HasSuffix(name, r.targets) == false {
		return nil, false, &core.TemplateError{Message: "template: \"" + name + "\" not defined"}
	}

	// ファイルを読み込む
	file, err := common.ReadFile(fmt.Sprintf("%s/%s", r.directory, name))
	// ファイルが存在しない、またはパーミッション等の理由でファイル読み込みが出来ない場合は、エラーとする
	if err != nil {
		return nil, false, &core.TemplateError{Message: "template: " + err.Error()}
	}
	defer file.Close()

	// バイナリファイルを対象としていない場合、エラーとする
	isBinary := file.IsBinary()
	if r.binary == false && isBinary {
		return nil, isBinary, &core.TemplateError{Message: "template: \"" + name + "\" not defined"}
	}

	// ファイルサイズ設定値を超過していた場合、エラーを返却する
	if r.maxsize > 0 && file.Size() > r.maxsize {
		return nil, false, &core.TemplateError{
			Message: fmt.Sprintf("%s: %d < %d. maxsize over", name, r.maxsize, file.Size()),
		}
	}

	return file.ReadAll(), isBinary, nil
}

// テンプレートオブジェクトを作成する
func (r *Render) template(name string, buf []byte, data interface{}) (tmpl *Template, err error) {
	tmpl = &Template{
		funcs: make(template.FuncMap),
	}

	// 一旦ヘルパ関数をコピーする
	for k, v := range r.funcs {
		tmpl.funcs[k] = v
	}
	// import : format で指定したテンプレートファイル名を元に、テンプレートファイルの内容をロードする
	tmpl.funcs["import"] = func(format string, i ...interface{}) (string, error) {
		// テンプレート名を変数へ格納
		var tmplname = format
		if len(i) >= 1 {
			tmplname = fmt.Sprintf(format, i...)
		}
		buf, err := r.execute(tmpl, tmplname, data)
		return string(buf), err
	}
	// hastemplate : 指定したテンプレート名が存在するかチェックする
	tmpl.funcs["hastemplate"] = func(format string, i ...interface{}) bool {
		// テンプレート名を変数へ格納
		var tmplname = format
		if len(i) >= 1 {
			tmplname = fmt.Sprintf(format, i...)
		}
		f, err := os.Stat(r.directory + "/" + tmplname)
		if err != nil || f.IsDir() {
			return false
		}
		return true
	}

	// ヘルパ関数をテンプレートオブジェクトへ登録する
	tmpl.Template, err = template.New(name).Funcs(tmpl.funcs).Parse(string(buf))
	if err != nil {
		return nil, common.RenderError(err, nil, string(buf))
	}
	return tmpl, nil
}

// パースしたテンプレートデータを実行解析する
func (r *Render) execute(tmpl *Template, name string, data interface{}) ([]byte, error) {
	// テンプレート情報をExecuteTemplateで解析し、結果をバッファヘ格納する
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, data)

	if err != nil {
		// エラーが発生した場合、RenderErrorか否かを判定する
		rerr := common.RenderError(err, tmpl.Template, "")
		if e, ok := rerr.(*core.RenderError); ok {
			// RenderError 型の場合は retry する
			if err := r.retry(tmpl, e.Target, rerr); err != nil {
				return nil, err
			}
			// retry 成功時は、再度executeを実行
			return r.execute(tmpl, name, data)
		}

		// エラー内容が、template: no template "..." not defined でないか確認する
		regex := regexp.MustCompile(`^template: no template "([^"]*?)" .*`)
		// ファイルが存在しない場合、リトライを試みる
		if regex.MatchString(err.Error()) {
			target := regex.ReplaceAllString(err.Error(), "$1")
			e := r.retry(tmpl, target, err)
			if e != nil {
				// retry 失敗理由がパースエラーの場合、エラーを返却
				if _, ok := e.(*core.RenderError); ok {
					return nil, e
				}
				// ファイル読み込み等のエラーの場合、エラーを返却
				return nil, &core.TemplateError{Message: e.Error()}
			}
			// retry 成功時は、再度executeを実行
			return r.execute(tmpl, name, data)
		}
	}
	// ExecuteTemplate成功の場合は、バッファに格納した情報を返却する
	return common.Exclude(buf.String(), r.exclude), nil
}

func (r *Render) retry(tmpl *Template, target string, err error) error {
	// ファイルを読み込む。失敗した場合は、元のエラーを返却する
	buf, isBinary, e := r.readfile(target)
	if e != nil {
		// tmpl.errors = append(tmpl.errors, err)
		return err
	}
	// 読み込んだデータがバイナリの場合はエラーを返却する
	if isBinary {
		return err
	}
	// ファイルデータをパースする。パース失敗時は、パースエラー内容を返却する
	v, err := tmpl.New(target).Parse(string(buf))
	if err != nil {
		return common.RenderError(err, tmpl.Template, string(buf))
	}
	tmpl.Template = v
	return nil
}

// CreateRender : レンダーオブジェクトを生成する
func CreateRender(c *common.Config) core.Render {
	return &Render{
		directory: c.Directory,
		targets:   c.Targets,
		exclude:   c.Exclude,
		binary:    c.Binary,
		maxsize:   c.MaxSize,
		funcs:     make(template.FuncMap),
	}
}
