package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BinokelPresenter ビノクルプレゼンターインタフェース
type BinokelPresenter interface {
	GamePresenter[interfaces.BinokelGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.BinokelGame) string
}
