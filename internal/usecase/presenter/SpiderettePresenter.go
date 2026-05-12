package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SpiderettePresenter スパイダレットプレゼンターインタフェース
type SpiderettePresenter interface {
	GamePresenter[interfaces.SpideretteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SpideretteGame) string
}
