package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BriscolaPresenter ブリスコラプレゼンターインタフェース
type BriscolaPresenter interface {
	GamePresenter[interfaces.BriscolaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BriscolaGame) string
}
