//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DilotiPresenter はディロティのプレゼンターインタフェース。
type DilotiPresenter interface {
	GamePresenter[interfaces.DilotiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.DilotiGame) string
}
