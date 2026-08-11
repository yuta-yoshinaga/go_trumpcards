//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RamsPresenter ラムスプレゼンターインタフェース
type RamsPresenter interface {
	GamePresenter[interfaces.RamsGame]
	// HintOutput ヒント情報を出力する
	HintOutput(r interfaces.RamsGame) string
}
