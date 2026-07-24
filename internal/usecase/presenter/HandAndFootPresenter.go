//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HandAndFootPresenter ハンドアンドフットプレゼンタインタフェース
type HandAndFootPresenter interface {
	GamePresenter[interfaces.HandAndFootGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.HandAndFootGame) string
}
