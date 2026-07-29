//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BraidPresenter ブレイド プレゼンターインタフェース
type BraidPresenter interface {
	GamePresenter[interfaces.BraidGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BraidGame) string
}
