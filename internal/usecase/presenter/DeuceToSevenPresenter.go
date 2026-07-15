//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DeuceToSevenPresenter is the 2-7 Triple Draw presenter alias over the generic
// GamePresenter. Adapter-side implementations satisfy this via Output +
// ActionLogOutput.
type DeuceToSevenPresenter interface {
	GamePresenter[interfaces.DeuceToSevenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.DeuceToSevenGame) string
}
