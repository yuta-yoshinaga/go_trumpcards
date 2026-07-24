//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// Rummy500Presenter Rummy 500プレゼンターインタフェース
type Rummy500Presenter interface {
	GamePresenter[interfaces.Rummy500Game]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.Rummy500Game) string
}
