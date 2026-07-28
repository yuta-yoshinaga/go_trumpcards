//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BouillottePresenter はブイヨット (Bouillotte) のプレゼンターインタフェース。
type BouillottePresenter interface {
	GamePresenter[interfaces.BouillotteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BouillotteGame) string
}
