//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// QuodlibetPresenter はクオドリベットのプレゼンターインタフェース。
type QuodlibetPresenter interface {
	GamePresenter[interfaces.QuodlibetGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.QuodlibetGame) string
}
