//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SimpleSimonPresenter シンプル・サイモンのプレゼンターインタフェース。
type SimpleSimonPresenter interface {
	GamePresenter[interfaces.SimpleSimonGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.SimpleSimonGame) string
}
