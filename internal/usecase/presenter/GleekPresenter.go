//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GleekPresenter グリーク (Gleek) のプレゼンターインタフェース
type GleekPresenter interface {
	GamePresenter[interfaces.GleekGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GleekGame) string
}
