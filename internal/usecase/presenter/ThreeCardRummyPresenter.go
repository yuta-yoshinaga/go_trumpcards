//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ThreeCardRummyPresenter スリーカード・ラミープレゼンターインタフェース
type ThreeCardRummyPresenter interface {
	GamePresenter[interfaces.ThreeCardRummyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ThreeCardRummyGame) string
}
