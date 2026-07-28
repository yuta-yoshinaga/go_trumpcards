//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ScartoPresenter スカルト (Scarto) のプレゼンターインタフェース
type ScartoPresenter interface {
	GamePresenter[interfaces.ScartoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ScartoGame) string
}
