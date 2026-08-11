//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GermanWhistPresenter ジャーマンホイストプレゼンターインタフェース
type GermanWhistPresenter interface {
	GamePresenter[interfaces.GermanWhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.GermanWhistGame) string
}
