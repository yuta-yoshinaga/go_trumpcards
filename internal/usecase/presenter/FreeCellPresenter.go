package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FreeCellPresenter フリーセルプレゼンターインタフェース
type FreeCellPresenter interface {
	GamePresenter[interfaces.FreeCellGame]
	// HintOutput ヒント情報を出力する
	HintOutput(f interfaces.FreeCellGame) string
}
