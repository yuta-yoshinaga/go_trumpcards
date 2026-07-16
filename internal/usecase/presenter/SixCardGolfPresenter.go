package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SixCardGolfPresenter シックスカードゴルフプレゼンターインタフェース
type SixCardGolfPresenter interface {
	GamePresenter[interfaces.SixCardGolfGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SixCardGolfGame) string
}
