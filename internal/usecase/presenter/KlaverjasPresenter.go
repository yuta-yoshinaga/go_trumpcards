//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KlaverjasPresenter クラヴァヤスのプレゼンターインタフェース
type KlaverjasPresenter interface {
	GamePresenter[interfaces.KlaverjasGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.KlaverjasGame) string
}
