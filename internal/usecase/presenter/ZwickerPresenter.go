//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ZwickerPresenter ツヴィッカープレゼンターインタフェース
type ZwickerPresenter interface {
	GamePresenter[interfaces.ZwickerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.ZwickerGame) string
}
