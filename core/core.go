package core

import "text/template"

// Render : Renderインタフェース
type Render interface {
	// Render オブジェクトをコピーする
	Copy() Render

	// 指定したヘルパ関数を所持しているか確認する
	HasHelper(string) bool

	// ヘルパを登録する。登録するヘルパは、構造体型を指定する必要がある
	Helper(interface{}) error
	LargeHelper(interface{}) error
	SmallHelper(interface{}) error
	AddHelper(template.FuncMap) error

	// 文字列のレンダーを処理する
	RenderString(string, interface{}) ([]byte, error)

	// 指定したレンダー名で、テンプレート解析を実施する
	Render(string, interface{}) ([]byte, error)
}

// HelperInvalid : ヘルパ登録時のエラー型
type HelperInvalid struct {
	Message string // エラーメッセージ本文
	Type    string // 構造体の型名
	Kind    string // Kind名
}

func (err *HelperInvalid) Error() string {
	return err.Message
}

// RenderError : Parse, Execute でエラーが起こった場合のエラー型
type RenderError struct {
	Message  string // エラーメッセージ
	Line     int    // エラー発生行番号
	Column   int    // エラーカラム番号
	Basename string // エラー元のレンダーファイル名
	Root     string // エラー元のレンダーファイル本文
	Target   string // 対象となるレンダーファイル名
}

func (err *RenderError) Error() string {
	return err.Message
}

// TemplateError : 存在しないテンプレートファイルを指定した場合のエラー
type TemplateError struct {
	Message string
}

func (err *TemplateError) Error() string {
	return err.Message
}
