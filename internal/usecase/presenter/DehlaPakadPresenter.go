//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DehlaPakadPresenter はデーラ・パカドのプレゼンターインタフェース。
type DehlaPakadPresenter interface {
	GamePresenter[interfaces.DehlaPakadGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.DehlaPakadGame) string
}
