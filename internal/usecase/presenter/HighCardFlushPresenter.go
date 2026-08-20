//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HighCardFlushPresenter ハイカードフラッシュプレゼンターインタフェース
type HighCardFlushPresenter interface {
	GamePresenter[interfaces.HighCardFlushGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.HighCardFlushGame) string
}
