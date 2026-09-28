//go:build !js || !wasm || extra9

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// StalactitesPresenter フリーセルプレゼンターインタフェース
type StalactitesPresenter interface {
	GamePresenter[interfaces.StalactitesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(f interfaces.StalactitesGame) string
}
