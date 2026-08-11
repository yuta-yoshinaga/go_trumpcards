//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PolignacPresenter ポリニャックプレゼンターインタフェース
type PolignacPresenter interface {
	GamePresenter[interfaces.PolignacGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.PolignacGame) string
}
