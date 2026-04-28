package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NertzPresenter Nertz / Pounce プレゼンターインタフェース
type NertzPresenter interface {
	GamePresenter[interfaces.NertzGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.NertzGame) string
}
