//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PokerSquaresPresenter はポーカー・スクエアズのプレゼンターインタフェース。
type PokerSquaresPresenter interface {
	GamePresenter[interfaces.PokerSquaresGame]
	// HintOutput 現在のカードを置く最善のセルのヒントを出力する
	HintOutput(p interfaces.PokerSquaresGame) string
}
