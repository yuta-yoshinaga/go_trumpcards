//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SlobberhannesPresenter スロバーハンネスプレゼンターインタフェース
type SlobberhannesPresenter interface {
	GamePresenter[interfaces.SlobberhannesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SlobberhannesGame) string
}
