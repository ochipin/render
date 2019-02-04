package render

import "testing"

func Test_CONFIG_NEW_ERROR(t *testing.T) {
	var conf = &Config{
		Directory:  "/nodir",
		Targets:    []string{".text", ".html", ".png"},
		Exclude:    nil,
		Cache:      true,
		Binary:     false,
		MaxSize:    10 << 20,
		SumMaxSize: 3000,
	}
	// 存在しないディレクトリを指定したため、エラーとなる
	if _, err := conf.New(); err == nil {
		t.Fatal("Error")
	}

	conf.Directory = "config.go"
	// ディレクトリではないファイルを指定したため、エラーとなる
	if _, err := conf.New(); err == nil {
		t.Fatal("Error")
	}

	// カレントディレクトリを対象とする
	conf.Directory = ""
	// 合計ファイルサイズが、SumMaxSizeを超過しているため、エラーとなる
	if _, err := conf.New(); err == nil {
		t.Fatal("Error")
	}

	conf.SumMaxSize = 0
	conf.MaxSize = 100
	// 単体ファイルサイズが、MaxSizeを超過しているため、エラーとなる
	if _, err := conf.New(); err == nil {
		t.Fatal("Error")
	}
}

func Test_CONFIG_NEW_SUCCESS(t *testing.T) {
	var conf = &Config{
		Directory:  "internal",
		Targets:    []string{".text", ".html", ".png"},
		Exclude:    nil,
		Cache:      true,
		Binary:     true,
		MaxSize:    5 << 20,
		SumMaxSize: 500 << 20,
	}
	// キャッシュありの場合
	if _, err := conf.New(); err != nil {
		t.Fatal("Error")
	}
	// キャッシュなしの場合
	conf.Cache = false
	if _, err := conf.New(); err != nil {
		t.Fatal("Error")
	}
}
