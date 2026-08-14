//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ColoradoPresenter コロラド プレゼンターインタフェース
type ColoradoPresenter interface {
	GamePresenter[interfaces.ColoradoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.ColoradoGame) string
}
