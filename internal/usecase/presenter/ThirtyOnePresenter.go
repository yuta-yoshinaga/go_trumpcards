//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ThirtyOnePresenter ThirtyOne プレゼンターインタフェース
// ThirtyOnePresenter は 31 のプレゼンターインタフェース。
type ThirtyOnePresenter interface {
	GamePresenter[interfaces.ThirtyOneGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ThirtyOneGame) string
}
