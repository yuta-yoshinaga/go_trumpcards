package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PageOnePresenter ページワンプレゼンターインタフェース
type PageOnePresenter interface {
	GamePresenter[interfaces.PageOneGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.PageOneGame) string
}
