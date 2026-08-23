//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BauernschnapsenPresenter バウエルンシュナプセンプレゼンターインタフェース
type BauernschnapsenPresenter interface {
	GamePresenter[interfaces.BauernschnapsenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BauernschnapsenGame) string
}
