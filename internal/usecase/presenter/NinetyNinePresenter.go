package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NinetyNinePresenter ナインティナインプレゼンターインタフェース
type NinetyNinePresenter interface {
	GamePresenter[interfaces.NinetyNineGame]
	// HintOutput ヒント情報を出力する
	HintOutput(o interfaces.NinetyNineGame) string
}
