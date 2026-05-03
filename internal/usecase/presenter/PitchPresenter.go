package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PitchPresenter ピッチプレゼンターインタフェース
type PitchPresenter interface {
	GamePresenter[interfaces.PitchGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.PitchGame) string
}
