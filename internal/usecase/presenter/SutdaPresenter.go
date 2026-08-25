//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SutdaPresenter はソッタのプレゼンターインタフェース。
type SutdaPresenter interface {
	GamePresenter[interfaces.SutdaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SutdaGame) string
}
