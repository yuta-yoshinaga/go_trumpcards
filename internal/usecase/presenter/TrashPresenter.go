package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TrashPresenter トラッシュプレゼンターインタフェース
type TrashPresenter interface {
	GamePresenter[interfaces.TrashGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.TrashGame) string
}
