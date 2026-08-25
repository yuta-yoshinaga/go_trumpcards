//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CostlyColoursPresenter はコストリー・カラーズのプレゼンターインタフェース。
type CostlyColoursPresenter interface {
	GamePresenter[interfaces.CostlyColoursGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CostlyColoursGame) string
}
