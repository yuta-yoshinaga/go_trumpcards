//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KingPresenter はキング (King) のプレゼンターインタフェース。
type KingPresenter interface {
	GamePresenter[interfaces.KingGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.KingGame) string
}
