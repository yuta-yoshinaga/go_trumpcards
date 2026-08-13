//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SakuraPresenter はさくら (肥後花) のプレゼンターインタフェース。
type SakuraPresenter interface {
	GamePresenter[interfaces.SakuraGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SakuraGame) string
}
