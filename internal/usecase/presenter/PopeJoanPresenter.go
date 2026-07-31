//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PopeJoanPresenter ポープ・ジョーンプレゼンターインタフェース
type PopeJoanPresenter interface {
	GamePresenter[interfaces.PopeJoanGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.PopeJoanGame) string
}
