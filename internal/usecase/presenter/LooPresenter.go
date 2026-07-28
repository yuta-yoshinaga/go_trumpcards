//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LooPresenter はルー (Loo) のプレゼンターインタフェース。
type LooPresenter interface {
	GamePresenter[interfaces.LooGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.LooGame) string
}
