//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TeenDoPaanchPresenter 3-2-5 プレゼンターインタフェース
type TeenDoPaanchPresenter interface {
	GamePresenter[interfaces.TeenDoPaanchGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TeenDoPaanchGame) string
}
