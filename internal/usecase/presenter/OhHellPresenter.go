package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OhHellPresenter オー・ヘルプレゼンターインタフェース
type OhHellPresenter interface {
	GamePresenter[interfaces.OhHellGame]
	// HintOutput ヒント情報を出力する
	HintOutput(o interfaces.OhHellGame) string
}
