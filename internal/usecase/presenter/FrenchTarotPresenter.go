//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FrenchTarotPresenter フレンチタロット (French Tarot) のプレゼンターインタフェース
type FrenchTarotPresenter interface {
	GamePresenter[interfaces.FrenchTarotGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.FrenchTarotGame) string
}
