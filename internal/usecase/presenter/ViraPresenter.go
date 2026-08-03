//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ViraPresenter ヴィーラのプレゼンターインタフェース
type ViraPresenter interface {
	GamePresenter[interfaces.ViraGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ViraGame) string
}
