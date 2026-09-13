//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// IndianRummyPresenter インドラミープレゼンターインタフェース
type IndianRummyPresenter interface {
	GamePresenter[interfaces.IndianRummyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.IndianRummyGame) string
}
