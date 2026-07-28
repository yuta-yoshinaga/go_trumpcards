//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BelotePresenter ベロートプレゼンターインタフェース
type BelotePresenter interface {
	GamePresenter[interfaces.BeloteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BeloteGame) string
}
