//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PrimeroPresenter はプリメロ (Primero) のプレゼンターインタフェース。
type PrimeroPresenter interface {
	GamePresenter[interfaces.PrimeroGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.PrimeroGame) string
}
