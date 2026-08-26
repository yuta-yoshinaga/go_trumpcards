//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MadrassoPresenter マドラッソのプレゼンターインタフェース
type MadrassoPresenter interface {
	GamePresenter[interfaces.MadrassoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MadrassoGame) string
}
