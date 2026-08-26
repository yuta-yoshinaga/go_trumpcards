//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BrusquembillePresenter ブリュスカンビーユプレゼンターインタフェース
type BrusquembillePresenter interface {
	GamePresenter[interfaces.BrusquembilleGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BrusquembilleGame) string
}
