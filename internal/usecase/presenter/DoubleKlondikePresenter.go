//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DoubleKlondikePresenter ダブル・クロンダイクのプレゼンターインタフェース。
type DoubleKlondikePresenter interface {
	GamePresenter[interfaces.DoubleKlondikeGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.DoubleKlondikeGame) string
}
