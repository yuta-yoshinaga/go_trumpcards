package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TrucoPresenter トゥルコプレゼンターインタフェース
type TrucoPresenter interface {
	GamePresenter[interfaces.TrucoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.TrucoGame) string
}
