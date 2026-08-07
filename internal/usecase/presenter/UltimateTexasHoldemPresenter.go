//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// UltimateTexasHoldemPresenter アルティメット・テキサスホールデムプレゼンターインタフェース
type UltimateTexasHoldemPresenter interface {
	GamePresenter[interfaces.UltimateTexasHoldemGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.UltimateTexasHoldemGame) string
}
