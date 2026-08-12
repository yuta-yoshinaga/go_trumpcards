//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LingerLongerPresenter リンガーロンガープレゼンターインタフェース
type LingerLongerPresenter interface {
	GamePresenter[interfaces.LingerLongerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.LingerLongerGame) string
}
