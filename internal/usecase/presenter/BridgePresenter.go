//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BridgePresenter ブリッジプレゼンターインタフェース
type BridgePresenter interface {
	GamePresenter[interfaces.BridgeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BridgeGame) string
}
