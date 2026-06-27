//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ThreeCardPresenter スリーカードポーカープレゼンターインタフェース
type ThreeCardPresenter interface {
	GamePresenter[interfaces.ThreeCardGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ThreeCardGame) string
}
