//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CribbageSquaresPresenter はクリベッジ・スクエアズのプレゼンターインタフェース。
type CribbageSquaresPresenter interface {
	GamePresenter[interfaces.CribbageSquaresGame]
	// HintOutput 現在のカードを置く最善のセルのヒントを出力する
	HintOutput(p interfaces.CribbageSquaresGame) string
}
