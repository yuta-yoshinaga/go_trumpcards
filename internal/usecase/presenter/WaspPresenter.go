package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WaspPresenter ワスププレゼンターインタフェース
type WaspPresenter interface {
	GamePresenter[interfaces.WaspGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.WaspGame) string
}
