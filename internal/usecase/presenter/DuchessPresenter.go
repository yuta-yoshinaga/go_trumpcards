//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DuchessPresenter ダッチェス プレゼンターインタフェース
type DuchessPresenter interface {
	GamePresenter[interfaces.DuchessGame]
	// HintOutput ヒント情報を出力する
	HintOutput(d interfaces.DuchessGame) string
}
