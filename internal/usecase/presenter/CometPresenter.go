//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CometPresenter はコメットのプレゼンターインタフェース。
type CometPresenter interface {
	GamePresenter[interfaces.CometGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CometGame) string
}
