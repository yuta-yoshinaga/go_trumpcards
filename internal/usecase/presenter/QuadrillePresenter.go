//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// QuadrillePresenter カドリール (Quadrille) のプレゼンターインタフェース
type QuadrillePresenter interface {
	GamePresenter[interfaces.QuadrilleGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.QuadrilleGame) string
}
