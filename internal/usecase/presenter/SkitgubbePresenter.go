//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SkitgubbePresenter シートグッベプレゼンターインタフェース
type SkitgubbePresenter interface {
	GamePresenter[interfaces.SkitgubbeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.SkitgubbeGame) string
}
