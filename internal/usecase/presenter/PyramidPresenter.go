//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PyramidPresenter ピラミッドプレゼンターインタフェース
type PyramidPresenter interface {
	GamePresenter[interfaces.PyramidGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.PyramidGame) string
}
