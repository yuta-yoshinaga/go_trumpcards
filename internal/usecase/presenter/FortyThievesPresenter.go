package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FortyThievesPresenter フォーティシーブスプレゼンターインタフェース
type FortyThievesPresenter interface {
	GamePresenter[interfaces.FortyThievesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(ft interfaces.FortyThievesGame) string
}
