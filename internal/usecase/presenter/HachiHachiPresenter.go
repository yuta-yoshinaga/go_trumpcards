//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HachiHachiPresenter は八八 (Hachi-Hachi) のプレゼンターインタフェース。
type HachiHachiPresenter interface {
	GamePresenter[interfaces.HachiHachiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.HachiHachiGame) string
}
