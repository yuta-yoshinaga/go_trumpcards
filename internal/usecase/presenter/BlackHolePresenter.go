//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BlackHolePresenter ブラックホールのプレゼンターインタフェース。
type BlackHolePresenter interface {
	GamePresenter[interfaces.BlackHoleGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.BlackHoleGame) string
}
