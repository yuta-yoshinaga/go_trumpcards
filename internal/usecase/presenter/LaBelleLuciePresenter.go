//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LaBelleLuciePresenter ラ・ベル・ルーシーのプレゼンターインタフェース。
type LaBelleLuciePresenter interface {
	GamePresenter[interfaces.LaBelleLucieGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.LaBelleLucieGame) string
}
