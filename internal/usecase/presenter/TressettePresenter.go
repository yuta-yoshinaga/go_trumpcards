//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TressettePresenter トレセッテのプレゼンターインタフェース
type TressettePresenter interface {
	GamePresenter[interfaces.TressetteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TressetteGame) string
}
