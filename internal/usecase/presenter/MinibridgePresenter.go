//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MinibridgePresenter ミニブリッジプレゼンターインタフェース
type MinibridgePresenter interface {
	GamePresenter[interfaces.MinibridgeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.MinibridgeGame) string
}
