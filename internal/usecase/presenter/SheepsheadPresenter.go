//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SheepsheadPresenter シープスヘッドのプレゼンターインタフェース
type SheepsheadPresenter interface {
	GamePresenter[interfaces.SheepsheadGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SheepsheadGame) string
}
