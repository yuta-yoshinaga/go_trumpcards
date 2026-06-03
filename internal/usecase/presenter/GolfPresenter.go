//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GolfPresenter ゴルフソリティアプレゼンターインタフェース
type GolfPresenter interface {
	GamePresenter[interfaces.GolfGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GolfGame) string
}
