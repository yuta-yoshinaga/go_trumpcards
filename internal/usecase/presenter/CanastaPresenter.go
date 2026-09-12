//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CanastaPresenter カナスタプレゼンタインタフェース
type CanastaPresenter interface {
	GamePresenter[interfaces.CanastaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CanastaGame) string
}
