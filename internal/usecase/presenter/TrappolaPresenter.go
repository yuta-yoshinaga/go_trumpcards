//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TrappolaPresenter トラッポラのプレゼンターインタフェース
type TrappolaPresenter interface {
	GamePresenter[interfaces.TrappolaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TrappolaGame) string
}
