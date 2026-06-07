//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EuchrePresenter ユーカープレゼンターインタフェース
type EuchrePresenter interface {
	GamePresenter[interfaces.EuchreGame]
	// HintOutput ヒント情報を出力する
	HintOutput(e interfaces.EuchreGame) string
}
