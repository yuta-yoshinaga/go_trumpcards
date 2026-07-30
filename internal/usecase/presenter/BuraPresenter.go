//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BuraPresenter ブラプレゼンターインタフェース
type BuraPresenter interface {
	GamePresenter[interfaces.BuraGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BuraGame) string
}
