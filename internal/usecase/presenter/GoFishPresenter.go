package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GoFishPresenter Go Fishプレゼンターインタフェース
type GoFishPresenter interface {
	GamePresenter[interfaces.GoFishGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.GoFishGame) string
}
