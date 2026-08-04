//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KaiserPresenter カイザー (Kaiser) プレゼンターインタフェース
type KaiserPresenter interface {
	GamePresenter[interfaces.KaiserGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.KaiserGame) string
}
