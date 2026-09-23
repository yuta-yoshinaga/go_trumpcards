//go:build !js || !wasm || extra5

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OmiPresenter オミプレゼンターインタフェース
type OmiPresenter interface {
	GamePresenter[interfaces.OmiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(e interfaces.OmiGame) string
}
