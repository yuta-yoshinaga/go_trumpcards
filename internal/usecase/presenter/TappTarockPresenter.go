//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TappTarockPresenter タップ・タロックのプレゼンターインタフェース。
type TappTarockPresenter interface {
	GamePresenter[interfaces.TappTarockGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TappTarockGame) string
}
