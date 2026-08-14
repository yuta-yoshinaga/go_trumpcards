//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RollingStonePresenter ローリングストーンプレゼンターインタフェース
type RollingStonePresenter interface {
	GamePresenter[interfaces.RollingStoneGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.RollingStoneGame) string
}
