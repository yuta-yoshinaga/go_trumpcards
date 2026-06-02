package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FiveHundredPresenter 500 (Five Hundred) プレゼンターインタフェース
type FiveHundredPresenter interface {
	GamePresenter[interfaces.FiveHundredGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.FiveHundredGame) string
}
