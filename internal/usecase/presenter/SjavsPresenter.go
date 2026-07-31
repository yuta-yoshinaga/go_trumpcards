//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SjavsPresenter シャウスプレゼンターインタフェース
type SjavsPresenter interface {
	GamePresenter[interfaces.SjavsGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.SjavsGame) string
}
