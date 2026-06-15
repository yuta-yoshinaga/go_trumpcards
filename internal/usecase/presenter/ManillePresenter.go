//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ManillePresenter マニーユのプレゼンターインタフェース
type ManillePresenter interface {
	GamePresenter[interfaces.ManilleGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ManilleGame) string
}
