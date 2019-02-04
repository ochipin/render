package render

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ochipin/render/internal/nocache"

	"github.com/ochipin/render/core"
	"github.com/ochipin/render/internal/cache"
	"github.com/ochipin/render/internal/common"
)

// Render : core.Render のエイリアス
type Render = core.Render

// Config : Renderインタフェースを構築する設定用構造体
type Config struct {
	Directory  string         // レンダー対象ディレクトリパス
	Targets    []string       // レンダー対象となるファイルの拡張子
	Exclude    *regexp.Regexp // レンダーファイル内の除外文字列
	Cache      bool           // true = オンメモリ, false = ディスク
	Binary     bool           // true = バイナリも扱う, false = バイナリは扱わない
	MaxSize    int64          // レンダーファイル1つにつき、最大で扱えるファイルサイズ
	SumMaxSize int64          // レンダーファイルの合計最大サイズ(Cache = true の時のみ有効)
}

// New : Renderインタフェースを生成する
func (config *Config) New() (Render, error) {
	var result core.Render

	// ディレクトリが未指定の場合、カレントディレクトリを対象とする
	if config.Directory == "" {
		config.Directory = "."
	}
	// 指定したパスが存在しない、またはディレクトリではない場合、エラーとする
	f, err := os.Stat(config.Directory)
	if err != nil {
		return nil, fmt.Errorf("cannot access '%s' no such file or directory", config.Directory)
	}
	if f.IsDir() == false {
		return nil, fmt.Errorf("cannot access '%s' not directory", config.Directory)
	}

	if config.Cache {
		// オンメモリの場合、キャッシュファイルリストを生成
		filelist, err := config.cacheFilelist()
		if err != nil {
			return nil, err
		}
		// レンダーオブジェクトを生成
		result = cache.CreateRender(&common.Config{
			Directory: strings.TrimRight(config.Directory, "/"),
			Exclude:   config.Exclude,
			Files:     filelist,
		})
	} else {
		// ディスクの場合
		result = nocache.CreateRender(&common.Config{
			Directory: strings.TrimRight(config.Directory, "/"),
			Targets:   config.Targets,
			Exclude:   config.Exclude,
			MaxSize:   config.MaxSize,
			Binary:    config.Binary,
		})
	}

	return result, nil
}

// Directory に指定したパス直下にある全ファイル一覧を取得し、レンダーファイルの元データを作成する
func (config *Config) cacheFilelist() ([]*common.File, error) {
	var filelist []*common.File
	var sumfilesize int64

	// 指定されたディレクトリ直下にあるファイル一覧を取得する
	err := filepath.Walk(config.Directory, func(path string, f os.FileInfo, err error) error {
		// ディレクトリの場合はスルー
		if f.IsDir() {
			return nil
		}
		// 登録済みの拡張子と一致しない場合は、スルー
		if common.HasSuffix(path, config.Targets) == false {
			return nil
		}
		// ファイルを読み込む
		file, err := common.ReadFile(path)
		// パーミッション等の理由でファイルが読み込み出来ない場合は、エラーとする
		if err != nil {
			return err
		}
		defer file.Close()
		// バイナリファイルを対象としていない場合、スルー
		isBinary := file.IsBinary()
		if config.Binary == false && isBinary == true {
			return nil
		}
		// ファイルサイズが設定値を超過していた場合、エラーを返却する
		if config.MaxSize > 0 && f.Size() > config.MaxSize {
			return fmt.Errorf("%s: %d < %d. maxsize over", path, config.MaxSize, f.Size())
		}
		// ファイルリストに、取得したファイル情報を追加
		filelist = append(filelist, &common.File{
			FileData: file.ReadAll(),
			FileName: common.ToWindowsPath(path[len(config.Directory)+1:]),
			IsBinary: isBinary,
		})
		// ファイルサイズの合計値を求める
		sumfilesize += f.Size()
		return nil
	})
	// ファイルサイズの合計値が、設定値であるSumMaxSizeを超過していないかチェック
	if config.SumMaxSize > 0 && sumfilesize > config.SumMaxSize {
		return nil, fmt.Errorf("%s: %d < %d. sum maxsize over", config.Directory, config.SumMaxSize, sumfilesize)
	}

	// ファイル一覧を返却する
	return filelist, err
}
