//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ScoponePresenter スコポーネプレゼンターインタフェース。
type ScoponePresenter interface {
	GamePresenter[interfaces.ScoponeGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ScoponeGame) string
}
