//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AluettePresenter アリュエットのプレゼンターインタフェース
type AluettePresenter interface {
	GamePresenter[interfaces.AluetteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.AluetteGame) string
}
