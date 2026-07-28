//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// UltiPresenter ウルティ (Ulti) のプレゼンターインタフェース
type UltiPresenter interface {
	GamePresenter[interfaces.UltiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.UltiGame) string
}
