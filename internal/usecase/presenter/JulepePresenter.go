//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// JulepePresenter フレペプレゼンターインタフェース
type JulepePresenter interface {
	GamePresenter[interfaces.JulepeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(r interfaces.JulepeGame) string
}
