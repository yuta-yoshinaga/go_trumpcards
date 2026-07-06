//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GoStopPresenter はゴーストップ (Go-Stop) のプレゼンターインタフェース。
type GoStopPresenter interface {
	GamePresenter[interfaces.GoStopGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GoStopGame) string
}
