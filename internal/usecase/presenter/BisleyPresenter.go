//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BisleyPresenter ビズリー プレゼンターインタフェース
type BisleyPresenter interface {
	GamePresenter[interfaces.BisleyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BisleyGame) string
}
