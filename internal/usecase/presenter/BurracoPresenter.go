//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BurracoPresenter ブラーコプレゼンタインタフェース
type BurracoPresenter interface {
	GamePresenter[interfaces.BurracoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BurracoGame) string
}
