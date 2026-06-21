//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ThreeCardBragPresenter スリーカード・ブラグのプレゼンターインタフェース
type ThreeCardBragPresenter interface {
	GamePresenter[interfaces.ThreeCardBragGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ThreeCardBragGame) string
}
