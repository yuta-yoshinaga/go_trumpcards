package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SpadesPresenter スペードプレゼンターインタフェース
type SpadesPresenter interface {
	GamePresenter[interfaces.SpadesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SpadesGame) string
}
