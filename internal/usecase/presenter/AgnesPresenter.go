//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AgnesPresenter アグネス・ソレルプレゼンターインタフェース
type AgnesPresenter interface {
	GamePresenter[interfaces.AgnesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.AgnesGame) string
}
