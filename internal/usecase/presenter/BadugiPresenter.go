//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BadugiPresenter is the Badugi-specific presenter alias over the generic
// GamePresenter. Adapter-side implementations satisfy this via Output +
// ActionLogOutput.
type BadugiPresenter interface {
	GamePresenter[interfaces.BadugiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BadugiGame) string
}
