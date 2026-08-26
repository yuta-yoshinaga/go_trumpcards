//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CoinchePresenter コワンシュプレゼンターインタフェース
type CoinchePresenter interface {
	GamePresenter[interfaces.CoincheGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.CoincheGame) string
}
