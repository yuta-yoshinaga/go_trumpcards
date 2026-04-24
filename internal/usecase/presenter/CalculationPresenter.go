package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CalculationPresenter カルキュレーションプレゼンターインタフェース
type CalculationPresenter interface {
	GamePresenter[interfaces.CalculationGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CalculationGame) string
}
