//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MichiganPresenter はミシガン (Michigan) のプレゼンターインタフェース。
type MichiganPresenter interface {
	GamePresenter[interfaces.MichiganGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MichiganGame) string
}
