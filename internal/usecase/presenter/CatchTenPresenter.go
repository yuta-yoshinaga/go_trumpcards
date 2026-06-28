package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CatchTenPresenter Catch the Ten プレゼンターインタフェース
type CatchTenPresenter interface {
	GamePresenter[interfaces.CatchTenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CatchTenGame) string
}
