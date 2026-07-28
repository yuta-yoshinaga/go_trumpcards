//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MightyPresenter マイティプレゼンターインタフェース
type MightyPresenter interface {
	GamePresenter[interfaces.MightyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(m interfaces.MightyGame) string
}
