package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SpiteAndMalicePresenter Spite & Malice プレゼンターインタフェース
type SpiteAndMalicePresenter interface {
	GamePresenter[interfaces.SpiteAndMaliceGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SpiteAndMaliceGame) string
}
