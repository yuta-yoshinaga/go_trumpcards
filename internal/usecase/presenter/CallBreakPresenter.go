package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CallBreakPresenter Call Break プレゼンターインタフェース
type CallBreakPresenter interface {
	GamePresenter[interfaces.CallBreakGame]
	// HintOutput ヒント情報を出力する
	HintOutput(cb interfaces.CallBreakGame) string
}
