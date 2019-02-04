package cache

import (
	"regexp"
	"sync"
	"text/template"

	"github.com/ochipin/render/core"
	"github.com/ochipin/render/internal/common"
)

// Render : キャッシュありのRenderオブジェクトを管理する構造体
type Render struct {
	mu       sync.Mutex
	filelist map[string]string
	binlist  map[string][]byte
	exclude  *regexp.Regexp
	funcs    template.FuncMap
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
		filelist: r.filelist,
		binlist:  r.binlist,
		exclude:  r.exclude,
		funcs:    funcs,
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
	// 所持しているレンダーファイルを解析する
	tmpl, err := r.template(data)
	if err != nil {
		return nil, err
	}
	// 渡された文字列ベースのテンプレートを解析
	tmpl, err = tmpl.New("string").Parse(text)
	if err != nil {
		return nil, common.RenderError(err, nil, text)
	}
	// 解析結果を返却する
	return common.Template(tmpl, "string", r.exclude, data)
}

// Render : 指定した名前でデータでテンプレートファイルの解析結果を取得する
func (r *Render) Render(tmplname string, data interface{}) ([]byte, error) {
	// バイナリファイルリストから指定された名前で登録されているバイナリファイルを取得する
	if v, ok := r.binlist[tmplname]; ok {
		return v, nil
	}
	// レンダーファイルリストから指定された名前で登録されているレンダーファイルを取得する
	if _, ok := r.filelist[tmplname]; !ok {
		return nil, &core.TemplateError{Message: "template: \"" + tmplname + "\" not defined"}
	}
	// レンダーファイルを解析する
	tmpl, err := r.template(data)
	if err != nil {
		return nil, err
	}

	// 解析結果を返却する
	return common.Template(tmpl, tmplname, r.exclude, data)
}

// テンプレートを解析
func (r *Render) template(data interface{}) (tmpl *template.Template, err error) {
	var funcs = make(template.FuncMap)

	// 一旦ヘルパ関数をコピーする
	for k, v := range r.funcs {
		funcs[k] = v
	}
	// import : format で指定したテンプレート名を元に、テンプレートファイルの内容をロードする
	funcs["import"] = func(format string, i ...interface{}) (string, error) {
		return common.Import(tmpl, data, format, i...)
	}
	// hastemplate : 指定したテンプレート名が存在するかチェックする
	funcs["hastemplate"] = func(format string, i ...interface{}) bool {
		return common.HasTemplate(tmpl, format, i...)
	}
	// レンダーファイルリストから、テンプレートを作成する
	for tmplname, tmpldata := range r.filelist {
		if tmpl == nil {
			tmpl, err = template.New(tmplname).Funcs(funcs).Parse(tmpldata)
		} else {
			tmpl, err = tmpl.New(tmplname).Parse(tmpldata)
		}
		// エラーが発生した場合、エラーを返却する
		if err != nil {
			return nil, common.RenderError(err, tmpl, tmpldata)
		}
	}

	return tmpl, err
}

// CreateRender : レンダーオブジェクトを生成する
func CreateRender(c *common.Config) core.Render {
	var filelist = make(map[string]string)
	var binlist = make(map[string][]byte)

	// ファイルリスト一覧の情報をもとに、バイナリ、レンダーファイルリストを作成する
	for _, v := range c.Files {
		if v.IsBinary {
			// バイナリファイルリストを作成
			binlist[v.FileName] = v.FileData
		} else {
			// レンダーファイルリストを作成
			filelist[v.FileName] = string(v.FileData)
		}
	}

	return &Render{
		filelist: filelist,
		binlist:  binlist,
		exclude:  c.Exclude,
		funcs:    make(template.FuncMap),
	}
}
