package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BakersDozenPresenter ベーカーズダズンプレゼンターインタフェース
type BakersDozenPresenter interface {
	GamePresenter[interfaces.BakersDozenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bd interfaces.BakersDozenGame) string
}
