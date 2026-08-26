//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AnacondaPresenter はアナコンダ (Anaconda) のプレゼンターインタフェース。
type AnacondaPresenter interface {
	GamePresenter[interfaces.AnacondaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.AnacondaGame) string
}
