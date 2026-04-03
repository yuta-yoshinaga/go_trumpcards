package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PinochlePresenter ピノクルプレゼンターインタフェース
type PinochlePresenter interface {
	GamePresenter[interfaces.PinochleGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.PinochleGame) string
}
