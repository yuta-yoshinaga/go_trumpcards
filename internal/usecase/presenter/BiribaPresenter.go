//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BiribaPresenter ビリバプレゼンタインタフェース
type BiribaPresenter interface {
	GamePresenter[interfaces.BiribaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BiribaGame) string
}
