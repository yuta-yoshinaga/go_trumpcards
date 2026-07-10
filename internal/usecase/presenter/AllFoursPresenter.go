package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AllFoursPresenter All Fours プレゼンターインタフェース
type AllFoursPresenter interface {
	GamePresenter[interfaces.AllFoursGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.AllFoursGame) string
}
