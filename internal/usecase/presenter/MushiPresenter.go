//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MushiPresenter 虫プレゼンターインタフェース
type MushiPresenter interface {
	GamePresenter[interfaces.MushiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(m interfaces.MushiGame) string
}
