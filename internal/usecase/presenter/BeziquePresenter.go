//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BeziquePresenter ベジークプレゼンターインタフェース
type BeziquePresenter interface {
	GamePresenter[interfaces.BeziqueGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BeziqueGame) string
}
