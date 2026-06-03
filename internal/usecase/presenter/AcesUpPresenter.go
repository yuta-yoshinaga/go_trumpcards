//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AcesUpPresenter エースアッププレゼンターインタフェース
type AcesUpPresenter interface {
	GamePresenter[interfaces.AcesUpGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.AcesUpGame) string
}
