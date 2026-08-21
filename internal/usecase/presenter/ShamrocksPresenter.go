//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ShamrocksPresenter シャムロックスのプレゼンターインタフェース。
type ShamrocksPresenter interface {
	GamePresenter[interfaces.ShamrocksGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.ShamrocksGame) string
}
