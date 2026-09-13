//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GinRummyPresenter ジンラミープレゼンターインタフェース
type GinRummyPresenter interface {
	GamePresenter[interfaces.GinRummyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GinRummyGame) string
}
