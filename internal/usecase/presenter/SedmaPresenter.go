//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SedmaPresenter セドマのプレゼンターインタフェース
type SedmaPresenter interface {
	GamePresenter[interfaces.SedmaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SedmaGame) string
}
