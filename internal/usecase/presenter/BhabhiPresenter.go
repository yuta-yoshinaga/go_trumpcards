//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BhabhiPresenter バービープレゼンターインタフェース
type BhabhiPresenter interface {
	GamePresenter[interfaces.BhabhiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BhabhiGame) string
}
