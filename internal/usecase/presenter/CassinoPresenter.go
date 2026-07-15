package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CassinoPresenter カシノプレゼンターインタフェース。
type CassinoPresenter interface {
	GamePresenter[interfaces.CassinoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(cg interfaces.CassinoGame) string
}
