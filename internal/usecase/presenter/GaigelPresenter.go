//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GaigelPresenter ガイゲルプレゼンターインタフェース
type GaigelPresenter interface {
	GamePresenter[interfaces.GaigelGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GaigelGame) string
}
