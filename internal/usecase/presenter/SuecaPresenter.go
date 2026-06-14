//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SuecaPresenter スエカのプレゼンターインタフェース
type SuecaPresenter interface {
	GamePresenter[interfaces.SuecaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SuecaGame) string
}
