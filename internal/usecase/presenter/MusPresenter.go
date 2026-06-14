//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MusPresenter ムスのプレゼンターインタフェース
type MusPresenter interface {
	GamePresenter[interfaces.MusGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MusGame) string
}
