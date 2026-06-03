//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GapsPresenter はGapsゲームのプレゼンターインタフェース。
type GapsPresenter interface {
	GamePresenter[interfaces.GapsGame]
	// HintOutput はヒント情報を出力する。
	HintOutput(g interfaces.GapsGame) string
}
