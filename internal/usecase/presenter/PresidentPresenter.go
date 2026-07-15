package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PresidentPresenter プレジデントプレゼンターインタフェース
type PresidentPresenter interface {
	GamePresenter[interfaces.PresidentGame]
	// HintOutput ヒント情報を出力する
	HintOutput(pg interfaces.PresidentGame) string
}
