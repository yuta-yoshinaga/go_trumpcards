//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PochPresenter ポッホプレゼンターインタフェース
type PochPresenter interface {
	GamePresenter[interfaces.PochGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.PochGame) string
}
