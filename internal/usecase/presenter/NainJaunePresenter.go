//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NainJaunePresenter ル・ナン・ジョーヌプレゼンターインタフェース
type NainJaunePresenter interface {
	GamePresenter[interfaces.NainJauneGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.NainJauneGame) string
}
