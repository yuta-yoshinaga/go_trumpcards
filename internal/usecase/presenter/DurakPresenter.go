package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DurakPresenter ドゥラークプレゼンターインタフェース
type DurakPresenter interface {
	GamePresenter[interfaces.DurakGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.DurakGame) string
}
