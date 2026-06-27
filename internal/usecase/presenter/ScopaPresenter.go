//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ScopaPresenter スコパプレゼンターインタフェース。
type ScopaPresenter interface {
	GamePresenter[interfaces.ScopaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ScopaGame) string
}
