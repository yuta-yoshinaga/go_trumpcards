//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AluettePresenter アリュエットのプレゼンターインタフェース
type AluettePresenter interface {
	GamePresenter[interfaces.AluetteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.AluetteGame) string
}
