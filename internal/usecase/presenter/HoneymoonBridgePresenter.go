//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HoneymoonBridgePresenter ハネムーンブリッジプレゼンターインタフェース
type HoneymoonBridgePresenter interface {
	GamePresenter[interfaces.HoneymoonBridgeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.HoneymoonBridgeGame) string
}
