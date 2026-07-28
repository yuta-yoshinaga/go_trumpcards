//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SevenBridgePresenter セブンブリッジプレゼンターインタフェース
type SevenBridgePresenter interface {
	GamePresenter[interfaces.SevenBridgeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SevenBridgeGame) string
}
