package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HeartsPresenter ハーツプレゼンターインタフェース
type HeartsPresenter interface {
	GamePresenter[interfaces.HeartsGame]
	// HintOutput ヒント情報を出力する
	HintOutput(h interfaces.HeartsGame) string
}
