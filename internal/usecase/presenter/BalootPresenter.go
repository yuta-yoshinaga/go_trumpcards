//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BalootPresenter バルートプレゼンターインタフェース
type BalootPresenter interface {
	GamePresenter[interfaces.BalootGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BalootGame) string
}
