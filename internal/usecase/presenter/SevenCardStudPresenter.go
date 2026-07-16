//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SevenCardStudPresenter セブンカードスタッドプレゼンターインタフェース
type SevenCardStudPresenter interface {
	GamePresenter[interfaces.SevenCardStudGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SevenCardStudGame) string
}
