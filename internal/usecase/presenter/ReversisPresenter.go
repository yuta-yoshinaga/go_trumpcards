//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ReversisPresenter レヴェルシプレゼンターインタフェース
type ReversisPresenter interface {
	GamePresenter[interfaces.ReversisGame]
	// HintOutput ヒント情報を出力する
	HintOutput(r interfaces.ReversisGame) string
}
