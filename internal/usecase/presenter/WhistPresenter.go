package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WhistPresenter ホイストプレゼンターインタフェース
type WhistPresenter interface {
	GamePresenter[interfaces.WhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(w interfaces.WhistGame) string
}
