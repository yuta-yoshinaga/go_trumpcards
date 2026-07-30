//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ToepenPresenter トゥーペンプレゼンターインタフェース
type ToepenPresenter interface {
	GamePresenter[interfaces.ToepenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.ToepenGame) string
}
