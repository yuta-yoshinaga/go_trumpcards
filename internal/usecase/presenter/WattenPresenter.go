//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WattenPresenter ヴァッテンプレゼンターインタフェース
type WattenPresenter interface {
	GamePresenter[interfaces.WattenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.WattenGame) string
}
