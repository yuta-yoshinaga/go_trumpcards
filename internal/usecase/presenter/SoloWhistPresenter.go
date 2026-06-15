//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SoloWhistPresenter ソロ・ホイストのプレゼンターインタフェース
type SoloWhistPresenter interface {
	GamePresenter[interfaces.SoloWhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SoloWhistGame) string
}
