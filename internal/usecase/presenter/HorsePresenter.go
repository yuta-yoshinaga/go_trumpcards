//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HorsePresenter は H.O.R.S.E. のプレゼンターインタフェース。
type HorsePresenter interface {
	GamePresenter[interfaces.HorseGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.HorseGame) string
}
