//go:build !js || !wasm || extra8

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BotifarraPresenter ボティファラプレゼンターインタフェース
type BotifarraPresenter interface {
	GamePresenter[interfaces.BotifarraGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.BotifarraGame) string
}
