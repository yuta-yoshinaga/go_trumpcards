package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AccordionPresenter アコーディオンプレゼンターインタフェース
type AccordionPresenter interface {
	GamePresenter[interfaces.AccordionGame]
	// HintOutput ヒント情報を出力する
	HintOutput(a interfaces.AccordionGame) string
}
