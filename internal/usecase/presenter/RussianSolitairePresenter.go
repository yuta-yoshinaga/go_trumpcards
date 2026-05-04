package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RussianSolitairePresenter ロシアンソリティアプレゼンターインタフェース
type RussianSolitairePresenter interface {
	GamePresenter[interfaces.RussianSolitaireGame]
	// HintOutput ヒント情報を出力する
	HintOutput(r interfaces.RussianSolitaireGame) string
}
