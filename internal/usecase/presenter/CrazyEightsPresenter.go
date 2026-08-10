package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CrazyEightsPresenter クレイジーエイトプレゼンターインタフェース
type CrazyEightsPresenter interface {
	GamePresenter[interfaces.CrazyEightsGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CrazyEightsGame) string
}
