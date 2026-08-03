//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GanjifaPresenter ガンジファのプレゼンターインタフェース
type GanjifaPresenter interface {
	GamePresenter[interfaces.GanjifaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GanjifaGame) string
}
