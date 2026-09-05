package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BatakPresenter Batak プレゼンターインタフェース
type BatakPresenter interface {
	GamePresenter[interfaces.BatakGame]
	// HintOutput ヒント情報を出力する
	HintOutput(cb interfaces.BatakGame) string
}
