//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ColourWhistPresenter カラーホイストプレゼンターインタフェース
type ColourWhistPresenter interface {
	GamePresenter[interfaces.ColourWhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.ColourWhistGame) string
}
