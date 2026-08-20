//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NapoleonPresenter ナポレオンプレゼンターインタフェース
type NapoleonPresenter interface {
	GamePresenter[interfaces.NapoleonGame]
	// HintOutput ヒント情報を出力する
	HintOutput(n interfaces.NapoleonGame) string
}
