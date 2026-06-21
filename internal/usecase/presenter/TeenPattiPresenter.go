//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TeenPattiPresenter ティーン・パティのプレゼンターインタフェース
type TeenPattiPresenter interface {
	GamePresenter[interfaces.TeenPattiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TeenPattiGame) string
}
